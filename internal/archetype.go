package internal

import (
	"fmt"
	"reflect"
	"slices"
)

// ArchetypeSignature is a unique identifier representing a specific
// group of component types.
type ArchetypeSignature []ComponentID

func (as ArchetypeSignature) String() string {
	return fmt.Sprintf("%v", as)
}

func MakeSignature(components ...Component) ArchetypeSignature {
	var compIds []ComponentID
	for _, c := range components {
		compIds = append(compIds, CompReg.GetComponentID(c))
	}
	slices.Sort(compIds)
	return ArchetypeSignature(compIds)
}

type Archetype struct {
	Signature      ArchetypeSignature
	EntityMap      map[EntityID]int             // Map an entity to its index in the component slice.
	Entities       []EntityID                   // List of entities matching this archetype.
	Components     map[ComponentID]any          // ComponentID->[]ComponentType
	ComponentTypes map[ComponentID]reflect.Type // ComponentID->Go Type
}

func NewArchetype(components ...Component) *Archetype {
	arch := &Archetype{
		EntityMap:      make(map[EntityID]int),
		Entities:       make([]EntityID, 0),
		Components:     make(map[ComponentID]any),
		ComponentTypes: make(map[ComponentID]reflect.Type),
	}

	var compIds []ComponentID
	for _, c := range components {
		id := CompReg.GetComponentID(c) // Get the comp ID for the component
		compIds = append(compIds, id)   // Append to the list

		compType := reflect.TypeOf(c)      // Get the Go type of the component
		arch.ComponentTypes[id] = compType // Map the ID->Go Type

		sliceType := reflect.SliceOf(compType)                               // Get the []<Component Type>
		arch.Components[id] = reflect.MakeSlice(sliceType, 0, 0).Interface() // Create a slice for the components

	}

	slices.Sort(compIds)
	arch.Signature = ArchetypeSignature(compIds)

	return arch
}

func (a *Archetype) AddEntity(entity EntityID, components ...Component) {
	idx := len(a.Entities)
	a.EntityMap[entity] = idx
	a.Entities = append(a.Entities, entity)

	for _, c := range components {
		id := CompReg.GetComponentID(c) // Lookup the id

		sliceVal := reflect.ValueOf(a.Components[id])           // Lookup the component slice
		sliceVal = reflect.Append(sliceVal, reflect.ValueOf(c)) // Append the component for the entity

		a.Components[id] = sliceVal.Interface()
	}
}

func (a *Archetype) HasComponents(query ArchetypeSignature) bool {
	archIdx, queryIdx := 0, 0

	// Linear scan over both the signature and query slices
	for archIdx < len(a.Signature) && queryIdx < len(query) {
		if a.Signature[archIdx] == query[queryIdx] {
			queryIdx++
		}
		archIdx++
	}

	// If we got through the entire query, it has all of the requisite components
	return queryIdx == len(query)
}
