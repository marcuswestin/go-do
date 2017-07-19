package do

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
)

type panicError struct {
	Stack   string
	Message string
}

func (p *panicError) Error() string {
	return p.Message
}
func RecoverPanic(err *error) {
	*err = recoverError()
}

func recoverError() error {
	recovery := recover()
	if recovery == nil {
		return nil
	} else {
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

		log.Printf("do.RecoverPanic recovered panic:\nError: \"%s\"\nStack: %s", msg, stack)
		bytes, _ := json.Marshal(&panicError{
			Stack:   string(debug.Stack()),
			Message: "Recovered panic: " + msg,
		})
		return errors.New(string(bytes))
	}
}
