package main

import (
  "fmt"
  "time"
  "math/rand"
  "os"
  "os/signal"
  "syscall"
  "net/http"
  "github.com/go-martini/martini"
)

type Bird struct {
  X int
  Y int
}

func (bird *Bird) Flap() {
  bird.Y -= 1
}

func (bird *Bird) Fall() {
  bird.Y += 1
}

func (bird *Bird) IsDead() bool {
  return bird.Y > 9
}

type Pipe struct {
  X int
  Height int
}

func (pipe *Pipe) ShiftLeft() {
  pipe.X -= 1
}

type Board struct {
  Pipes []Pipe
}

func (board *Board) RandomizePipes() {
  var currentX = 80
  for _,p := range board.Pipes {
    p.Height = rand.Intn(6) + 1
    p.X = currentX
    currentX += 7
  }
}

func (board *Board) ShiftPipes() {
  for _,p := range board.Pipes {
    p.ShiftLeft()
  }
}

type Game struct {
  Score int
}

func main() {
  os.Exit(realMain())
}

func realMain() int {
  sigs := make(chan os.Signal, 1)
  done := make(chan bool, 1)
  flapChannel := make(chan bool, 1)

  signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

  clear()

  go serverLoop(flapChannel)
  go gameLoop(flapChannel)

  go func() {
    <- sigs
    done <- true
  }()

  <-done
  fmt.Println("Exiting")
  return 0;
}

func clear() {
  //http.Get("http://10.105.4.251/peggy/clear?board=0")
  //http.Get("http://10.105.4.251/peggy/write?board=0&x=0&y=10&text={g}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}")
  //http.Get("http://10.105.4.251/peggy/write?board=0&x=0&y=11&text={g}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}")
}

func draw(bird Bird, board Board) {
  for _,p := range board.Pipes {
    fmt.Println("PIPE",p);
    if (p.X < 80) {
      //var pipeTop = "TTTTT"
      //var pipeMiddle = "%20{f}{f}{f}%20"
      // Top fo pipe = 11 - Ground Height - Height
      var topY = 11 - 2 - p.Height

      // Pipe is X
      // Previous Pipe drawing is X - 1
      // Draw pipe top (X - 1)
      //http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%20%20%20%20%20",p.X-1,topY))
      // Draw middles (X - 1)
      for y := topY+1; y < 10; y++ {
        //http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%20%20%20%20%20",p.X-1,y))
      }

      // Draw pipe top
      //http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%s",p.X,topY,pipeTop))
      // Draw middle * (height - 3)
      for y := topY+1; y < 10; y++ {
        //http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%s",p.X,y,pipeMiddle))
      }
    }
  }
  //http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%%20%%20%%20", bird.X, bird.Y+1))
  //http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%%20%%20%%20", bird.X, bird.Y-1))
  //http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text={r}~B~", bird.X, bird.Y))
}

func gameLoop(flaps <-chan bool) {
  died := make(chan bool, 1)

  bird := Bird{Y: 5, X: 35}
  board := Board{Pipes: make([]Pipe, 100) }
  board.RandomizePipes()
  game := Game{Score: 0}

  clear()

  for {
    select {
    case <- died:
      fmt.Println("Died...")
      //http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=20&y=5&text={r}SCORE:%%20%d", game.Score))
      time.Sleep(time.Second * 5)
    case <- flaps:
      bird.Flap()
    default:
      bird.Fall()
      if (bird.IsDead()) { died <- true }
      game.Score++
    }
    board.ShiftPipes()
    time.Sleep(time.Second * 2)
    draw(bird,board)
  }
}

func serverLoop(flaps chan<- bool) {
  m := martini.Classic()
  m.Post("/flap", func(w http.ResponseWriter, r *http.Request) {
    flaps <- true
    http.Redirect(w, r, "/", 302)
  })
  m.Run()
}
