package internal

import "reflect"

type Component any

type ComponentID uint

var (
	CompReg                     = NewComponentRegistry()
	nextComponentID ComponentID = 0
)

type ComponentRegistry struct {
	store map[reflect.Type]ComponentID
}

func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{store: make(map[reflect.Type]ComponentID)}
}

func (r *ComponentRegistry) GetComponentID(c Component) ComponentID {
	t := reflect.TypeOf(c)
	if id, ok := r.store[t]; ok {
		return id
	}

	id := nextComponentID
	r.store[t] = id
	nextComponentID++
	return id
}

type IdentityComponent struct {
	Name string
	Age  int // 0 - 100
}

type StatComponent struct {
	Health int // 0 - 100
	// Hunger
	// Energy
	// Stress
	// etc.
}

type MoodComponent struct {
	Happiness int // 0 - 100
	// Anger
	// etc.
}
