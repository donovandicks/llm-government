package internal

import (
	"context"
	"sync"
	"time"
)

type World struct {
	sync.RWMutex

	inputs map[string]Input

	clock time.Duration
	tick  int64
}

func (w *World) RegisterInput(in Input) {
	w.Lock()
	defer w.Unlock()

	w.inputs[in.Name()] = in
}

func (w *World) GetInput(name string) (Input, bool) {
	w.RLock()
	defer w.RUnlock()

	in, ok := w.inputs[name]
	return in, ok
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

	// For each agent
	// 	Observe the world state
	// 	Send observation to agent to get decision
	// 		Possible A2A discussion/debate/consensus?
	// 	Apply actions

	// For each system
	// 	Execute system

	// For each metric
	// 	Compute metric
}
