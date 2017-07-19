package do

import (
	"sync"
)

// ParallelWorkPool allows for executing work in parallel, while ensuring
// that the number of parallel executions are kept below a certain number.
type ParallelWorkPool interface {
	GetWork() (workItem interface{}, isDone bool)
	ReportWork(error)
	WaitCh() <-chan error
	Wait() error
}

type parallelWorkPool struct {
	promise     Promise
	workChannel chan interface{}
	poolChannel chan struct{}
	wg          sync.WaitGroup
}

// GetWork returns a work item as soon as the following
func (p *parallelWorkPool) GetWork() (workItem interface{}, isDone bool) {
	<-p.poolChannel
	i, open := <-p.workChannel
	return i, !open
}
func (p *parallelWorkPool) ReportWork(err error) {
	if err != nil {
		p.promise.Reject(err)
		return
	}
	p.poolChannel <- struct{}{}
	p.wg.Done()
}

func (p *parallelWorkPool) Wait() error          { return p.promise.Wait() }
func (p *parallelWorkPool) WaitCh() <-chan error { return p.promise.WaitCh() }

// NewParallelWorkPool creates a ParallelWorkPool and asyncronously pipes the contents from
// the input channel into the pool. You are responsible for closing the input channel when
// there is no more work to be done.
//
// The ParallelWorkPool allows you to consume and perform work in parallel, while ensuring
// that no more than numParallel work items are processed in parallel.
//
// To consume and perform work, call `workItem, isDone := p.GetWork()` until `isDone == true`.
// When `isDone == false`, you must call `p.ReportWork(err)` for each work item.
//
// 	pool := do.NewParallelWorkPool(ctx, workChannel, 5) // ensure at most 5 parallel processes
// 	for workItem := range pool.GetWork()
func NewParallelWorkPool(ctx Context, inputChannel <-chan interface{}, numParallel int) ParallelWorkPool {
	if numParallel < 1 {
		panic("ParallelWorkPool called with numParallel < 1")
	}
	promise := NewPromise()
	workChannel := make(chan interface{}, numParallel)
	poolChannel := make(chan struct{}, numParallel)
	wg := sync.WaitGroup{}
	workPool := &parallelWorkPool{promise, workChannel, poolChannel, wg}

	// Seed work pool - these are async since poolChannel is buffered
	for i := 0; i < numParallel; i++ {
		workPool.poolChannel <- struct{}{}
	}

	// Pipe work from input channel into parallel work channel
	go func() {
		for {
			select {
			case <-promise.WaitCh():
				// there has been an error. Stop.
				return

			case <-ctx.Done():
				// context has been cancelled. Stop.
				promise.Reject(ctx.Err())
				return

			case workItem, isOpen := <-inputChannel:
				if isOpen {
					// Received work item from input channel.
					// Increment wait counter, pipe work item into
					// parallel work pool, and continue.
					workPool.wg.Add(1)
					workPool.workChannel <- workItem
					continue

				} else {
					// input channel was closed, all input has been consumed.
					// Wait for all work to complete, and then resolve promise.
					close(workPool.workChannel)
					workPool.wg.Wait()
					promise.Resolve(nil)
					return
				}
			}
		}
	}()

	return workPool
}
