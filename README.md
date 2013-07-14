# Workout

Workout is a work processing library for [Go](http://golang.org) built on top of [beanstalkd](http://kr.github.io/beanstalkd). It provides a simple API, allowing you to get started quickly with reasonable default settings. Under the hood, Workout uses goroutines and a configurable number of workers to coordinate concurrent processing.

## Example

`
package main

import (
  "fmt"
  "github.com/jcoene/workout"
  "os"
  "os/signal"
  "time"
)

func main() {
  // These are the tubes we'll be using
  tubes := []string{"default", "person", "address"}

  // Create a new workout client for job insertion
  wc, _ := workout.NewClient("localhost:11300", tubes)

  // Insert 1000 jobs for each tube
  for i := 0; i < 1000; i++ {
    for _, t := range tubes {
      job := &workout.Job{
        Tube:      t,
        Priority:  1,
        TimeToRun: (60 * time.Second),
        Body:      fmt.Sprintf("%d", i),
      }
      wc.Put(job)
    }
  }

  // Setup a workout master with 20 workers
  wm := workout.NewMaster("localhost:11300", tubes, 20)

  // Assign our doNothing handler to each tube
  for _, t := range tubes {
    wm.RegisterHandler(t, doNothing)
  }

  // Start processing!
  wm.Start()

  // Tell workout to stop on CTRL+C
  go func() {
    ch := make(chan os.Signal)
    signal.Notify(ch, os.Interrupt)
    <-ch
    wm.Stop()
  }()

  // Block until we're finished
  wm.Wait()

  return
}

// A handler function takes a job pointer and returns an error (or nil on success)
func doNothing(job *workout.Job) (err error) {
  return nil
}
`

## License

MIT license, see [LICENSE](https://github.com/jcoene/workout/blob/master/LICENSE) for details.


## Author

Jason Coene, [@jcoene](https://twitter.com/jcoene)
