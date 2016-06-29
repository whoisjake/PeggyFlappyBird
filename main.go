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
  PreviousY int
  Body string
}

func (bird *Bird) Flap() {
  bird.PreviousY = bird.Y
  bird.Y -= 1
}

func (bird *Bird) Fall() {
  bird.PreviousY = bird.Y
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
  Pipes []*Pipe
}

func randomBoard() *Board {
  currentX := 80
  board := &Board{Pipes: make([]*Pipe, 100) }

  for i:=0; i < 100; i++ {
    board.Pipes[i] = &Pipe{Height: rand.Intn(6) + 1, X: currentX}
    currentX += 7
  }

  return board
}

func (board *Board) ShiftPipes() {
  for _,p := range board.Pipes {
    p.ShiftLeft()
  }
}

type Game struct {
  Score int
}

func write(x int, y int, content string) {
  http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%s",x,y,content))
}

func clear() {
  http.Get("http://10.105.4.251/peggy/clear?board=0")
  write(0,10,"{g}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}")
  write(0,11,"{g}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}{f}")
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



func draw(bird *Bird, board *Board) {
  for _,p := range board.Pipes {
    if (p.X < 80) {
      pipeTop := "TTTTT"
      pipeMiddle := "%20{f}{f}{f}%20"
      // Top fo pipe = 11 - Ground Height - Height
      topY := 11 - 2 - p.Height

      // Pipe is X
      // Previous Pipe drawing is X - 1
      // Draw pipe top (X - 1)
      write(p.X-1,topY,"%20%20%20%20%20")
      // Draw middles (X - 1)
      for y := topY+1; y < 10; y++ {
        write(p.X-1,y,"%20%20%20%20%20")
      }

      // Draw pipe top
      write(p.X,topY,pipeTop)

      // Draw middle * (height - 3)
      for y := topY+1; y < 10; y++ {
        write(p.X,y,pipeMiddle)
      }
    }
  }

  write(bird.X,bird.PreviousY,"%20%20%20")
  write(bird.X,bird.Y,bird.Body)
}

func gameLoop(flaps <-chan bool) {
  died := make(chan bool, 1)

  bird := &Bird{Y: 5, X: 35, Body: "{r}~B~"}
  board := randomBoard()
  game := &Game{Score: 0}

  clear()

  for {
    select {
    case <- died:
      fmt.Println("Died...")
      write(20,5,fmt.Sprintf("{r}SCORE:%%20%d",game.Score))
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
