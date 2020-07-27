package service

import (
	"time"
)

// PersonEntity represents the structure of a person in the database
type PersonEntity struct {
	ID         string
	firstName  string
	middleName string
	lastName   string
	suffix     string
	created    *time.Time
	updated    *time.Time
}

//Bind binds the value to the instance of the object
func (p *PersonEntity) Bind(ID string, firstName string, middleName string, lastName string, suffix string, created *time.Time, updated *time.Time) *PersonEntity {
	p.ID = ID
	p.firstName = firstName
	p.middleName = middleName
	p.lastName = lastName
	p.suffix = suffix
	p.created = created
	p.updated = updated
	return p
}
