package do_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	do "github.com/marcuswestin/go-x-do"
)

var ctx = context.Background()

func Example() {
	var workCount int32
	maxParallel := 5
	ctx := context.Background()

	// Parallel loop:
	items := make([]string, 50)
	loopCh := do.Async(func() error {
		return do.ParallelLoop(ctx, len(items), maxParallel,
			func(i int) error {
				defer atomic.AddInt32(&workCount, 1)
				// This function will be running in parallel,
				// with at most 5 concurrent executions.
				// fmt.Println("Item", i, items[i])
				return nil
			})
	})

	// Parallel channel read:
	ch := make(chan interface{})
	go func() {
		for i := 0; i < 50; i++ {
			line := fmt.Sprint("String ", i)
			ch <- line
		}
		close(ch)
	}()
	readCh := do.Async(func() error {
		return do.ParallelRead(ctx, ch, maxParallel, func(workItem interface{}) error {
			defer atomic.AddInt32(&workCount, 1)
			// line := workItem.(string)
			// fmt.Println("Do work on line: ", line)
			return nil
		})
	})

	// Parallel work generation and execution:
	workCh := do.Async(func() error {
		return do.ParallelWork(ctx, maxParallel,
			func(workCh chan<- interface{}) error {
				for i := 0; i < 50; i++ {
					line := fmt.Sprint("String ", i)
					workCh <- line
				}
				return nil
			},
			func(workItem interface{}) error {
				defer atomic.AddInt32(&workCount, 1)
				// line := workItem.(string)
				// fmt.Println("Do work on line: ", line)
				return nil
			})
	})
	// Wait for multiple error channels. Bails early if any of them has an error.
	err := do.WaitForErrorChannels(ctx, loopCh, readCh, workCh)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Println("Done work:", workCount)

	// Output: Done work: 150
}

func ExampleWaitForErrorChannels() {
	errCh1 := do.Async(func() error {
		time.Sleep(1 * time.Millisecond)
		fmt.Println("Hello 1")
		return nil
	})
	errCh2 := do.Async(func() error {
		time.Sleep(5 * time.Millisecond)
		fmt.Println("Hello 2")
		return nil
	})
	errCh3 := do.Async(func() error {
		time.Sleep(10 * time.Millisecond)
		return errors.New("Hello 3 error")
	})
	fmt.Println("Hello before")
	err := do.WaitForErrorChannels(ctx, errCh1, errCh2, errCh3)
	if strings.Contains(err.Error(), "Hello 3 error") {
		fmt.Println("Got Hello 3 error")
	}
	fmt.Println("Hello after")

	// Output: Hello before
	// Hello 1
	// Hello 2
	// Got Hello 3 error
	// Hello after
}

func ExampleRecoverPanic() {
	panicCh := do.Async(func() error {
		panic("Hello 1 panic")
	})

	panicFn := func() (err error) {
		defer do.RecoverPanic(&err)
		// ...
		panic("Hello 2 panic")
	}

	err := <-panicCh
	if strings.Contains(err.Error(), "Hello 1 panic") {
		fmt.Println("Got Hello 1 panic")
	}

	err = panicFn()
	if strings.Contains(err.Error(), "Hello 2 panic") {
		fmt.Print("Got Hello 2 panic")
	}

	// Output: Got Hello 1 panic
	// Got Hello 2 panic
}
