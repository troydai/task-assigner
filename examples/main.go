package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/troydai/task-assigner/internal/assigner"
	"github.com/troydai/task-assigner/internal/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	assigner := assigner.NewManager(ctx, NewProcessor())

	wg := &sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer wg.Done()

			t := StringTask(fmt.Sprintf("input-%v", idx))
			r, err := assigner.Assign(ctx, &t)
			if err != nil {
				log.Fatalf("unexpected err: %s", err)
			}
			fmt.Println("assigned task", t)

			result := <-r
			fmt.Printf("result of task %s = %s\n", t, result.Output)
		}(i)

	}

	wg.Wait()
}

type StringTask string

var _ types.Task = (*StringTask)(nil)

func (t StringTask) Key() types.TaskKey { return types.TaskKey(t) }

func (t StringTask) Input() types.InputData { return t }

type JitterStringProcessor struct {
	ch chan *jitterResult
}

type jitterResult struct {
	key   types.TaskKey
	value types.OutputData
}

var _ assigner.Processor = (*JitterStringProcessor)(nil)

func NewProcessor() *JitterStringProcessor {
	p := &JitterStringProcessor{
		ch: make(chan *jitterResult),
	}
	return p
}

func (p *JitterStringProcessor) In(key types.TaskKey, input types.InputData) error {
	wait := time.Duration(rand.Int31n(10)) * time.Second

	time.AfterFunc(wait, func() {
		p.ch <- &jitterResult{
			key:   key,
			value: fmt.Sprintf("[[%s]]", input),
		}
	})

	return nil
}

func (p *JitterStringProcessor) Out() (types.TaskKey, types.OutputData) {
	r := <-p.ch
	return r.key, r.value
}
