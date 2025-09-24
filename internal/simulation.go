package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type PopulationConfig struct {
	Size int `json:"size" yaml:"size"` // The amount of people in the scenario.
}

type Simulation struct {
	id         string            // Unique simulation ID generated at runtime, used for telemetry correlation.
	Scenario   string            `json:"scenario" yaml:"scenario"`                         // The scenario in which the agents are participating.
	Population *PopulationConfig `json:"population,omitempty" yaml:"population,omitempty"` // Details about the population in the scenario.
}

func LoadSimulationFromFile(filePath string) (*Simulation, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var unmarshaler func([]byte, any) error
	var sim Simulation
	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		unmarshaler = yaml.Unmarshal
	} else if strings.HasSuffix(filePath, ".json") {
		unmarshaler = json.Unmarshal
	} else {
		return nil, fmt.Errorf("cannot load simulation: unkown format")
	}

	if err := unmarshaler(data, &sim); err != nil {
		return nil, err
	}

	sim.id = uuid.NewString()
	return &sim, nil
}

func (s *Simulation) ID() string { return s.id }
