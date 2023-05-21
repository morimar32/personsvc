package graph

import (
	person "personsvc/generated"
	"personsvc/graph/model"
	"sync/atomic"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	persons    []*model.Person
	Client     person.PersonClient
	QueryCount *atomic.Int32
}
