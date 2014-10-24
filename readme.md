# Q

An in-memory (bounded, if that's your thing, because it probably should be) queue to use in Go programs.

```go
package main

import (
        "fmt"
        "os"
        "os/signal"
        "time"

        "github.com/danielscottt/q"
)

func h(event interface{}) {
        time.Sleep(time.Duration(event.(int)) * time.Millisecond)
        fmt.Println(event.(int))
}

func main() {

        sigc := make(chan os.Signal, 0)
        signal.Notify(sigc, os.Interrupt)
        queue := q.NewQueue(5)
        queue.Handler = h

        go queue.Work()

        // deliver some events
        for _, i := range []int{16, 8, 15, 4, 42, 23} {
                go func(i int) {
                        queue.Event <- i
                }(i)
        }

        for {
                select {
                case err := <-queue.Err:
                        fmt.Println(err.Error())
                case s := <-sigc:
                        fmt.Println(s, "received. Exiting...")
                        os.Exit(0)
                }
        }

}
```
