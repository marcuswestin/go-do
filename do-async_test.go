package do_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	do "github.com/marcuswestin/go-x-do"
)

var ctx = context.Background()

func ExampleAsyncParallel() {
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

func ExampleAsyncRecover() {
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
