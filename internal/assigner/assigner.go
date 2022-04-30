package assigner

import (
	"context"
	"sync"

	"github.com/troydai/task-assigner/internal/types"
)

type (
	Processor interface {
		In(key types.TaskKey, data types.InputData) error
		Out() (types.TaskKey, types.OutputData)
	}

	Manager struct {
		p     Processor
		mu    sync.Mutex
		tasks map[types.TaskKey]chan<- *types.Result
	}
)

var _ types.TaskAssigner = (*Manager)(nil)

func NewManager(ctx context.Context, p Processor) *Manager {
	m := &Manager{
		p:     p,
		tasks: map[types.TaskKey]chan<- *types.Result{},
		mu:    sync.Mutex{},
	}

	go func() {
		for {
			k, v := p.Out()

			ch := m.removeTask(k)
			if ch != nil {
				defer close(ch)
				ch <- &types.Result{
					Key:    k,
					Output: v,
				}
			}

			delete(m.tasks, k)
		}
	}()

	return m
}

func (m *Manager) removeTask(key types.TaskKey) chan<- *types.Result {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch, found := m.tasks[key]
	if found {
		delete(m.tasks, key)
	}

	return ch
}

func (m *Manager) Assign(ctx context.Context, task types.Task) (<-chan *types.Result, error) {
	key, input := task.Key(), task.Input()
	if err := m.p.In(key, input); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan *types.Result)
	m.tasks[key] = ch

	return ch, nil
}
