package internal

import (
	"math/rand/v2"

	"github.com/go-faker/faker/v4"
)

type EntityID uint

func NewPersonEntity() []Component {
	return []Component{
		IdentityComponent{
			Name: faker.Name(),
			Age:  rand.IntN(100),
		},
		StatComponent{
			Health: 100,
		},
		MoodComponent{
			Happiness: 100,
		},
	}
}
