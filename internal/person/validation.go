package service

import (
	br "github.com/morimar32/helpers/errors"
)

func (i *PersonHandler) validateUpdate(p *PersonEntity) error {
	if p == nil {
		return br.NewValidationError("request was empty")
	}
	if len(p.ID) <= 0 {
		return br.NewValidationError("Id is required")
	}

	return i.validateCommon(p)
}

func (i *PersonHandler) validateCommon(p *PersonEntity) error {
	if p == nil {
		return br.NewValidationError("request was empty")
	}
	if len(p.FirstName) <= 0 {
		return br.NewValidationError("First name is required")
	}
	if len(p.LastName) <= 0 {
		return br.NewValidationError("LastName is required")
	}
	if len(p.FirstName) > 50 {
		return br.NewValidationError("FirstName is too long. Maximum value is 50 characters")
	}
	if p.MiddleName != "" && len(p.MiddleName) > 50 {
		return br.NewValidationError("MiddleName is too long. Maximum value is 50 characters")
	}
	if len(p.LastName) > 100 {
		return br.NewValidationError("LastName is too long. Maximum value is 100 characters")
	}
	if p.Suffix != "" && len(p.Suffix) > 20 {
		return br.NewValidationError("Suffix is too long. Maximum value is 20 characters")
	}
	return nil
}
