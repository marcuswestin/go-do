package do

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

type Context context.Context
type GenWorkFn func(workCh chan<- interface{}) error
type DoWorkFn func(workItem interface{}) error
type LoopFn func(i int) error

// Async executes a function asyncronously, and returns a channel
// which will contain the functions return error value. If the
// function panics, Async recovers the panic and sends an error
// containing the panic error message and stack trace to the
// returned channel.
func Async(fn func() error) chan error {
	ch := make(chan error)
	go func() {
		defer func() {
			if err := recoverError(); err != nil {
				ch <- err
			}
		}()
		ch <- fn()
	}()
	return ch
}

// WaitForErrorChannels will wait to receive a value from each given channel.
// If any received value is a non-nil error, WaitForErrorChannels returns
// immediately with that given error value (and stops reading from all channels).
// If the given context is cancelled, WaitForErrorChannels returns immediately
// with the context.Err() value.
func WaitForErrorChannels(ctx Context, channels ...<-chan error) (err error) {
	cases := make([]reflect.SelectCase, len(channels)+1)
	ctxDoneCaseIndex := len(channels)
	for i, ch := range channels {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	if ctx != nil {
		cases[ctxDoneCaseIndex] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}
	}

	remaining := len(channels)
	for remaining > 0 {
		i, value, ok := reflect.Select(cases)

		// Catch error cases and return early
		if i == ctxDoneCaseIndex {
			// Context was cancelled
			return ctx.Err()

		} else if !value.IsNil() {
			// Channel received an error
			return value.Interface().(error)

		} else if !ok {
			// Channel was closed
			return errors.New(fmt.Sprintf("WaitForErrorChannels attempted read from closed channel #%d", i))
		}

		// Set this channel to nil so that we don't read from it twice.
		cases[i].Chan = reflect.ValueOf(nil)
		remaining -= 1
	}
	return nil
}

// NonBlockingRead does one non-blocking read from errChan. If there was a value to be
// read from channel it is returned, along with didRead=true; or else nil and didRead=false.
func NonBlockingRead(channel chan interface{}) (item interface{}, didRead bool) {
	return nonBlockingRead(channel)
}

// NonBlockingReadErr does one non-blocking read from errChan. If there was an error to be
// read from errChan it is returned, along with didRead=true; or else nil and didRead=false.
func NonBlockingReadErr(errChan chan error) (err error, didRead bool) {
	val, didRead := nonBlockingRead(errChan)
	return val.(error), didRead
}

// NonBlockingReadStruct does one non-blocking read from channel. If there was a message to be
// read from channel it returns didRead=true; or else didRead=false.
func NonBlockingReadStruct(channel chan struct{}) (didRead bool) {
	_, didRead = nonBlockingRead(channel)
	return didRead
}

func nonBlockingRead(channel interface{}) (item interface{}, didRead bool) {
	rCase := reflect.SelectCase{
		Chan: reflect.ValueOf(channel),
		Dir:  reflect.SelectRecv,
	}
	_, value, didRead := reflect.Select([]reflect.SelectCase{rCase})
	return value.Interface(), didRead
}
