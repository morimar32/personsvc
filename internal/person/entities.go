package service

import (
	"time"
)

// PersonEntity represents the structure of a person in the database
type PersonEntity struct {
	ID         string
	FirstName  string
	MiddleName string
	LastName   string
	Suffix     string
	Created    *time.Time
	Updated    *time.Time
}

//Bind binds the value to the instance of the object
func (p *PersonEntity) Bind(ID string, firstName string, middleName string, lastName string, suffix string, created *time.Time, updated *time.Time) *PersonEntity {
	p.ID = ID
	p.FirstName = firstName
	p.MiddleName = middleName
	p.LastName = lastName
	p.Suffix = suffix
	p.Created = created
	p.Updated = updated
	return p
}
