package main

import (
  "fmt"
  "time"
  "os"
  "os/signal"
  "syscall"
  "net/http"
  "github.com/go-martini/martini"
)

type board struct {
  bird_y_pos int
  bird_x_pos int
  column_heights []int
  rising bool
}

func main() {
  os.Exit(realMain())
}

func realMain() int {
  sigs := make(chan os.Signal, 1)
  done := make(chan bool, 1)
  flapChannel := make(chan bool, 1)

  signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

  http.Get("http://10.105.4.251/peggy/clear?board=0")
  http.Get("http://10.105.4.251/peggy/write?board=0&x=0&y=10&text={g}**********************************************************************************")
  http.Get("http://10.105.4.251/peggy/write?board=0&x=0&y=11&text={g}**********************************************************************************")

  go serverLoop(flapChannel)
  go gameLoop(flapChannel)

  go func() {
    sig := <-sigs
    fmt.Println()
    fmt.Println(sig)
    done <- true
  }()

  <-done
  fmt.Println("Exiting")
  return 0;
}

func draw(board *board) {
  if (board.rising) {
    http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%%20", board.bird_x_pos, board.bird_y_pos+1))
  } else {
    http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text=%%20", board.bird_x_pos, board.bird_y_pos-1))
  }
  http.Get(fmt.Sprintf("http://10.105.4.251/peggy/write?board=0&x=%d&y=%d&text={r}B", board.bird_x_pos, board.bird_y_pos))
}

func gameLoop(flaps <-chan bool) {
  died := make(chan bool, 1)

  bird_y_pos := 5
  bird_x_pos := 20
  columns := make([]int, 80)
  rising := false

  for {
    started := time.Now()
    select {
    case <- died:
      fmt.Println("Died...")
      bird_y_pos = 5
      columns = make([]int, 80)
    case <- flaps:
      bird_y_pos -= 1;
      rising = true
      if bird_y_pos < 0 { bird_y_pos = 0 }
      go func() {
        time.Sleep(time.Second * 3)
        rising = false
      }()
    default:
      fmt.Println(started)
      fmt.Println("Rising:", rising)
      fmt.Println("Bird Y Pos:", bird_y_pos)

      if (!rising) {
        bird_y_pos += 1
        if bird_y_pos > 10 { died <- true }
      }
      time.Sleep(time.Second * 2)
    }
    draw(&board{ bird_x_pos: bird_x_pos, bird_y_pos: bird_y_pos, column_heights: columns, rising: rising});
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
