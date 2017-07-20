// Package do gives you powerful asyncronous and parallel execution tooling.
//
// For API reference, see https://godoc.org/github.com/marcuswestin/go-x-do#pkg-index.
//
// For general example usage, see https://godoc.org/github.com/marcuswestin/go-x-do#example-package.
//
// For specific function examples and tests, see https://godoc.org/github.com/marcuswestin/go-x-do#pkg-examples.
package do

import (
	"context"
)

// Context is an alias for context.Context
type Context context.Context

// GenWorkFn is expected to fill a channel with work, and then return.
// The GenWorkFn should not close the channel. If the GenWorkFn returns
// an error, all subsequent work execution is stopped.
type GenWorkFn func(workCh chan<- interface{}) error

// WorkFn is expected to process work items, one at a time, and return
// when done. If a WorkFn returns an error, subsequent work execution
// is stopped.
type WorkFn func(workItem interface{}) error
