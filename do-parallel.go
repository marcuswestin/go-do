package do

// ParallelWork will call the genWorkFn asynchronously, and expects it to generate work items.
// It then calls the WorkFn once for each generated work item, and ensures that at most numParallel
// invocations of the work function are running in parallel. If any doWork invocation returns an
// error then ParallelWork will immediately return with that error and stop performing any new work.
func ParallelWork(ctx Context, numParallel int, genWorkFn GenWorkFn, doWorkFn WorkFn) error {
	workChan := make(chan interface{})
	genErrCh := Async(func() error {
		defer close(workChan)
		return genWorkFn(workChan)
	})
	readErrCh := Async(func() error {
		return ParallelRead(ctx, workChan, numParallel, doWorkFn)
	})

	return WaitForErrorChannels(ctx, genErrCh, readErrCh)
}

// ParallelRead will read values from channel until it is closed, and execute workFn once for
// each item read. There will be at most numParallel invocations of workFn at any given time.
// If any workFn returns a non-nil error, ParallelRead will stop reading values from channel
// and return the error (any errors of already invoked parallel workFns are simply ignored).
func ParallelRead(ctx Context, channel <-chan interface{}, numParallel int, workFn WorkFn) error {
	parallelWorkPool := NewParallelWorkPool(ctx, channel, numParallel)

	go func() {
		for {
			workItem, isDone := parallelWorkPool.GetWork()
			if isDone {
				return
			}
			go func() {
				err := workFn(workItem)
				parallelWorkPool.ReportWork(err)
			}()
		}
	}()
	return <-parallelWorkPool.WaitCh()
}

// ParallelLoop will call the given loopFn numItems times, with i ranging from
// 0 to numItems. There will be at most numParallel invocations of loopFn at any
// given time. If any loopFn invocation returns an error then ParallelLoop will
// immediately return with that error and stop performing any new loops.
func ParallelLoop(ctx Context, numItems, numParallel int, loopFn func(i int) error) error {
	loopCh := make(chan interface{})
	go func() {
		for i := 0; i < numItems; i++ {
			loopCh <- i
		}
		close(loopCh)
	}()
	return ParallelRead(ctx, loopCh, numParallel, func(item interface{}) error {
		return loopFn(item.(int))
	})
}
