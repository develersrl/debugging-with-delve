package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Handler is the type of a handler function to be executed in a separate
// goroutine managed by the worker.
type Handler func() (any, error)

// Worker represents a background worker that waits on the Start channel
// to execute a task in a separate goroutine.
type Worker struct {
	Start   <-chan struct{}
	Handler Handler
}

// NewWorker takes a handler to be execute by the worker and a channel
// to listen for a start signal, and returns a new Worker.
func NewWorker(h Handler, ch <-chan struct{}) *Worker {
	return &Worker{
		Start:   ch,
		Handler: h,
	}
}

// Run starts the background goroutine of the worker, taking a context
// and a waitgroup to support graceful shutdown.
func (w *Worker) Run(ctx context.Context, wg *sync.WaitGroup) (<-chan any, <-chan error) {
	results := make(chan any, 1)
	errs := make(chan error)

	go func() {
		defer wg.Done()
		defer close(errs)
		defer close(results)

		// wait for start
		<-w.Start

		for {
			select {
			case <-ctx.Done():
				return
			default:
				res, err := w.Handler()
				if err != nil {
					errs <- err
				} else {
					results <- res
				}
			}
		}
	}()

	return results, errs
}

func work() (any, error) {
	time.Sleep(250 * time.Millisecond)

	value := rand.Intn(100)

	if value >= 50 {
		return 0, errors.New("value greater than 50")
	}

	return value, nil
}

func main() {
	start := make(chan struct{})

	worker := NewWorker(work, start)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)
	results, errs := worker.Run(ctx, &wg)

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	start <- struct{}{}

out:
	for {
		select {
		case <-stop:
			// SIGINT received
			break out
		case res := <-results:
			fmt.Printf("work done: %v\n", res)
		case err := <-errs:
			fmt.Printf("work error: %v\n", err)
		}
	}

	cancel()

	wg.Wait()

	fmt.Println("goodbye!")
}
