package types

import "context"

type (
	TaskKey string

	InputData interface{}

	OutputData interface{}

	Task interface {
		Input() InputData
		Key() TaskKey
	}

	Result struct {
		Key    TaskKey
		Output OutputData
		Err    error
	}

	TaskAssigner interface {
		Assign(ctx context.Context, task Task) (<-chan *Result, error)
	}
)
