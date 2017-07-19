package do

// Package do gives you powerful asyncronous and parallel execution tooling.
//
// Example parallel execution usage:
//
// 	maxParallel := 5
// 	ctx := context.Background()
//
// 	// Parallel loop:
// 	items := make([]string, 50)
// 	loopCh := do.ParallelLoop(ctx, len(items), maxParallel,
// 	    func(i int) error {
// 	        // This function will be running in parallel,
// 	        // with at most 5 concurrent executions.
// 	        log.Println("Item", i, items[i])
// 	        return nil
// 	    })
//
// 	// Parallel channel read:
// 	ch := make(chan string)
// 	go func() {
// 	    for i := 0; i<50; i++ {
// 	        line := fmt.Sprint("String ", i)
// 	        ch <- line
// 	    }
// 	    close(ch)
// 	}()
// 	readCh := do.ParallelRead(ctx, ch, maxParallel,
// 	    func(workItem interface{}) error {
// 	        line := workItem.(string)
// 	        log.Println("Do work on line: ", line)
// 	        return nil
// 	    })
//
// 	// Parallel work generation and execution:
// 	workCh := do.ParallelWork(ctx, numParallel,
// 	    func(workCh chan<- interface{}) error {
// 	        for i := 0; i<50; i++ {
// 	            line := fmt.Sprint("String ", i)
// 	            workCh <- line
// 	        }
// 	        return nil
// 	    },
// 	    func(workItem interface{}) error {
// 	        line := workItem.(string)
// 	        log.Println("Do work on line: ", line)
// 	        return nil
// 	    })
//
// 	// Wait for multiple error channels. Bails early if any of them has an error.
// 	err := do.WaitForErrorChannels(loopCh, readCh, workCh)
// 	if err != nil {
// 	    log.Println("Error: ", err)
// 	}

import (
	"context"
)

type Context context.Context
type GenWorkFn func(workCh chan<- interface{}) error
type DoWorkFn func(workItem interface{}) error
type LoopFn func(i int) error
