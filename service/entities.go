package service

import "time"

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
