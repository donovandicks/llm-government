package internal

import "sync"

type Observation struct {
	Tick    int64
	Metrics map[string]any
	Inputs  map[string]any
	// Seed    int64
}

type Action struct {
	InputName string
	Value     any
}

type Input interface {
	Name() string
	Get() any
	Set(value any)
}

type SimpleInput struct {
	sync.RWMutex

	name  string
	value any
}

func NewSimpleInput(name string, initial any) *SimpleInput {
	return &SimpleInput{
		name:  name,
		value: initial,
	}
}

func (s *SimpleInput) Name() string { return s.name }

func (s *SimpleInput) Get() any {
	s.RLock()
	defer s.RUnlock()
	return s.value
}

func (s *SimpleInput) Set(value any) {
	s.Lock()
	defer s.Unlock()
	s.value = value
}
