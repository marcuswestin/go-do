package do

import (
	"fmt"
	"sync"
)

type Promise interface {
	Wait() error
	WaitCh() <-chan error
	Get() (resolvedValue interface{})
	Resolve(value interface{})
	Reject(err error)
	Check() (error, bool)
}

func NewPromise() Promise {
	return &promise{sync.Mutex{}, make(chan error), nil, nil, false}
}

type promise struct {
	mx  sync.Mutex
	ch  chan error
	val interface{}
	err error

	wasResolved bool
}

func (p *promise) Get() interface{} {
	if err := p.Wait(); err != nil {
		panic("Promise.Get() called on rejected promise: " + p.err.Error())
	}
	return p.val
}
func (p *promise) WaitCh() <-chan error {
	return p.ch
}
func (p *promise) Wait() error {
	return <-p.ch
}
func (p *promise) Check() (error, bool) {
	return CheckErrChan(p.ch)
}

// Resolve must only be called once. After Resolve has been called, any call to Reject will panic.
func (p *promise) Resolve(value interface{}) {
	p.fulfill(true, value, nil)
}

// Reject may be called multiple times, but must never be called after Resolve has been called.
func (p *promise) Reject(err error) {
	p.fulfill(false, nil, err)
}

func (p *promise) fulfill(isResolution bool, val interface{}, err error) {
	p.mx.Lock()
	defer p.mx.Unlock()

	// Ensure that Resolve or Reject are never called on an already
	// resolved promise.
	if p.wasResolved {
		if isResolution {
			panic(fmt.Sprintf("Promise.Resolve() called twice. First: %v. Second: %v", p.val, val))
		} else {
			panic(fmt.Sprintf("Promise.Resolve() was followed by Promise.Reject(). Resolution: %v. Rejection: %v", p.val, err))
		}
	}
	// Ensure that Resolve is never called after Reject.
	// Multiple calls to Reject are OK though - only the first one is recorded.
	if p.err != nil {
		if isResolution {
			panic(fmt.Sprintf("Promise.Reject() was followed by Promise.Resolve(). Rejection: %v. Resolution: %v", p.err, val))
		} else {
			// multiple calls to Reject are OK.
			return
		}
	}

	// Record promise fulfillment
	p.wasResolved = isResolution
	p.val = val
	p.err = err

	// Allow for any number of calls to Wait() or reads from WaitCh()
	go func() {
		for {
			p.ch <- err
		}
	}()
}
