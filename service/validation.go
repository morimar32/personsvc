package service

import (
	"fmt"

	person "personsvc/generated"
)

func (s *PersonService) validateGet(req *person.PersonRequest) error {
	if req == nil {
		return fmt.Errorf("request was empty")
	}
	if len(req.Id) <= 0 {
		return fmt.Errorf("Id is required")
	}
	return nil
}

func (s *PersonService) validateUpdate(req *person.UpdatePersonRequest) error {
	if req == nil {
		return fmt.Errorf("request was empty")
	}
	if len(req.Id) <= 0 {
		return fmt.Errorf("Id is required")
	}
	if len(req.FirstName) <= 0 {
		return fmt.Errorf("First name is required")
	}
	if len(req.LastName) <= 0 {
		return fmt.Errorf("LastName is required")
	}
	if len(req.FirstName) > 50 {
		return fmt.Errorf("FirstName is too long. Maximum value is 50 characters")
	}
	if req.MiddleName != nil && len(req.MiddleName.Value) > 0 && len(req.MiddleName.Value) > 50 {
		return fmt.Errorf("MiddleName is too long. Maximum value is 50 characters")
	}
	if len(req.LastName) > 100 {
		return fmt.Errorf("LastName is too long. Maximum value is 100 characters")
	}
	if req.Suffix != nil && len(req.Suffix.Value) > 20 {
		return fmt.Errorf("Suffix is too long. Maximum value is 20 characters")
	}
	return nil
}
