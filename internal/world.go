package internal

import (
	"context"
	"sync"
	"time"
)

type Observation struct {
	Tick      int64                  `json:"tick"`
	Timestamp int64                  `json:"timestamp"`
	Inputs    map[string]any         `json:"inputs"`
	Outputs   map[string]OutputValue `json:"outputs"`
	// Seed    int64
}

type World struct {
	sync.RWMutex

	inputs  map[string]Input
	outputs map[string]Output

	clock time.Duration
	tick  int64
}

func NewWorld() *World {
	return &World{
		inputs:  make(map[string]Input),
		outputs: make(map[string]Output),
	}
}

func (w *World) RegisterInput(in Input) *World {
	w.Lock()
	defer w.Unlock()

	w.inputs[in.Name()] = in
	return w
}

func (w *World) GetInput(name string) (Input, bool) {
	w.RLock()
	defer w.RUnlock()

	in, ok := w.inputs[name]
	return in, ok
}

func (w *World) RegisterOutput(out Output) *World {
	w.Lock()
	defer w.Unlock()

	w.outputs[out.Name()] = out
	return w
}

func (w *World) GetOutput(name string) (Output, bool) {
	w.RLock()
	defer w.RUnlock()

	out, ok := w.outputs[name]
	return out, ok
}

// Tick processes the world one dt at a time.
//
// Example:
// ```go
// tickDuration := time.Second
//
//	for i := 0; i < 20; i++ {
//		world.Tick(ctx, tickDuration)
//	}
//
// ```
func (w *World) Tick(ctx context.Context, dt time.Duration) {
	w.tick++
	w.clock += dt

	// For each system
	// 	Execute system
}

func (w *World) Observe(ctx context.Context) Observation {
	w.RLock()
	defer w.RUnlock()

	inputs := make(map[string]any)
	for name, in := range w.inputs {
		inputs[name] = in.Get()
	}

	outputs := make(map[string]OutputValue)
	for name, out := range w.outputs {
		outputs[name] = out.Compute(ctx, w)
	}

	return Observation{
		Tick:      w.tick,
		Timestamp: time.Now().UnixMilli(),
		Inputs:    inputs,
		Outputs:   outputs,
	}
}
