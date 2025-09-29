package internal

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"reflect"

	"github.com/go-faker/faker/v4"
)

type EntityID uint

type Person struct {
	IdentityComponent
	StatComponent
	MoodComponent
}

func NewPersonEntity() []Component {
	return []Component{
		IdentityComponent{
			Name: fmt.Sprintf("%s %s", faker.FirstName(), faker.LastName()),
			Age:  rand.IntN(100),
		},
		StatComponent{
			Health: 100,
			Money:  rand.IntN(200000),
		},
		MoodComponent{
			Happiness: 100,
		},
	}
}

func (qr *QueryResult) ToPersons() []Person {
	persons := make([]Person, 0, qr.Count)
	for personIdx := range qr.Count {
		person := Person{}
		for _, component := range qr.Components {
			switch v := component.(type) {
			case []IdentityComponent:
				person.Name = v[personIdx].Name
				person.Age = v[personIdx].Age
			case []StatComponent:
				person.Health = v[personIdx].Health
				person.Money = v[personIdx].Money
			case []MoodComponent:
				person.Happiness = v[personIdx].Happiness
			default:
				slog.Warn("invalid type for person component", "type", reflect.TypeOf(component))
			}
		}
		persons = append(persons, person)
	}

	return persons
}
