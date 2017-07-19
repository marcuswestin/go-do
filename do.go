package do

import (
	"context"
)

type Context context.Context
type GenWorkFn func(workCh chan<- interface{}) error
type DoWorkFn func(workItem interface{}) error
type LoopFn func(i int) error
