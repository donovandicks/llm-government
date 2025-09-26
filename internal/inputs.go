package internal

import "sync"

type Action struct {
	InputName string
	Value     any
}

type Input interface {
	Name() string
	Description() string
	Get() any
	Set(value any)
}

type SimpleInput struct {
	sync.RWMutex

	name        string
	description string
	value       any
}

func NewSimpleInput(name, description string, initial any) *SimpleInput {
	return &SimpleInput{
		name:        name,
		description: description,
		value:       initial,
	}
}

func (s *SimpleInput) Name() string        { return s.name }
func (s *SimpleInput) Description() string { return s.description }

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
