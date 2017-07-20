package do

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
)

// Async executes a function asyncronously, and returns a channel
// which will contain the functions return error value. If the
// function panics, Async recovers the panic and sends an error
// containing the panic error message and stack trace to the
// returned channel.
func Async(fn func() error) chan error {
	ch := make(chan error)
	go func() {
		defer func() {
			if err := recoverError(recover()); err != nil {
				ch <- err
			}
		}()
		ch <- fn()
	}()
	return ch
}

// RecoverPanic will recover from panics, and will assign the given *error pointer
// with an informative error, including the panic message and a stack trace. You
// should call RecoverPanic with defer, and pass it a pointer to the function's
// error return value. Example usage:
//
// 	func DoSomething() (err error) {
// 		defer do.RecoverPanic(&err)
//		...
// 		// The following panic will not bubble up to the called of DoSomething.
// 		// Instead, DoSomething will return an error with a message that
// 		// contains "Oh no!!", as well as a stack trace to this line.
// 		panic("Oh no!!")
// 	}
func RecoverPanic(err *error) {
	*err = recoverError(recover())
}

// LogPanics - if true, do.Async and do.RecoverPanic will log.Printf an informative
// debug message for every recovered panic, including the panic message and stack trace.
var LogPanics = false

type panicError struct {
	Stack   string
	Message string
}

func (p *panicError) Error() string {
	return p.Message
}

func recoverError(recovery interface{}) error {
	if recovery == nil {
		return nil
	}
	var msg string
	switch recovery := recovery.(type) {
	case error:
		msg = recovery.Error()
	default:
		msg = fmt.Sprint(recovery)
	}
	stack := string(debug.Stack())
	lines := strings.Split(stack, "\n")
	stack = lines[0] + "\n[ ... debug.Stack, do.RecoverPanic, panic ...]\n" + strings.Join(lines[7:], "\n")

	if LogPanics {
		log.Printf("do.RecoverPanic recovered panic:\nError: \"%s\"\nStack: %s", msg, stack)
	}
	bytes, _ := json.Marshal(&panicError{
		Stack:   string(debug.Stack()),
		Message: "Recovered panic: " + msg,
	})
	return errors.New(string(bytes))
}
