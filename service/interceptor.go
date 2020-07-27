package service

import (
	"context"
	"sync"
	"sync/atomic"

	br "github.com/morimar32/helpers/errors"
)

type PersonInterceptor struct {
	db IPersonRepository
}

func NewPersonInterceptor(repo *IPersonRepository) PersonInterceptor {
	val := &PersonInterceptor{
		db: *repo,
	}
	return *val
}

var (
	getPersonEntity = sync.Pool{
		New: func() interface{} {
			return &PersonEntity{}
		},
	}
	entityAllocCount int64
)

// GetPersonEntity Retrieves a PersonEntity from the pool
func GetPersonEntity() *PersonEntity {
	entity := getPersonEntity.Get().(*PersonEntity)
	atomic.AddInt64(&entityAllocCount, 1)
	return entity
}

// PutPersonEntity Returns a PersonEntity to the pool
func PutPersonEntity(entity *PersonEntity) {
	getPersonEntity.Put(entity)
	atomic.AddInt64(&entityAllocCount, -1)
}

// GetPerson returns an instance of a Person
func (p *PersonInterceptor) GetPerson(ctx context.Context, id string) (*PersonEntity, error) {
	if len(id) <= 0 {
		return nil, br.NewValidationError("Id is required")
	}

	model, err := p.db.Get(ctx, id)
	if err != nil {
		return nil, br.NewDataAccessErrorFromError(err)
	}
	if model == nil {
		return nil, br.NewValidationError("Person not found")
	}

	return model, nil
}

// GetPersons returns a list of people
func (p *PersonInterceptor) GetPersons(ctx context.Context) ([]*PersonEntity, error) {
	vals, err := p.db.GetList(ctx)
	if err != nil {
		return nil, br.NewDataAccessErrorFromError(err)
	}
	if vals == nil {
		return nil, br.NewValidationError("No Persons found")
	}
	return vals, nil
}

// AddPerson creates a new person record in the system
func (p *PersonInterceptor) AddPerson(ctx context.Context, add *PersonEntity) (*PersonEntity, error) {
	err := p.validateCommon(add)
	if err != nil {
		return nil, err
	}

	person, err := p.db.Add(ctx, add)
	if err != nil {
		return nil, br.NewDataAccessErrorFromError(err)
	}
	return person, nil
}

// UpdatePerson updates an existing person record in the system
func (p *PersonInterceptor) UpdatePerson(ctx context.Context, update *PersonEntity) (*PersonEntity, error) {
	err := p.validateUpdate(update)
	if err != nil {
		return nil, err
	}
	person, err := p.db.Update(ctx, update)
	if err != nil {
		return nil, br.NewDataAccessErrorFromError(err)
	}
	return person, nil
}

// DeletePerson removes a person from the system
func (p *PersonInterceptor) DeletePerson(ctx context.Context, id string) (bool, error) {
	if len(id) < 36 {
		return false, br.NewValidationError("Invalid id")
	}
	deleted, err := p.db.Delete(ctx, id)
	if err != nil {
		return false, br.NewDataAccessErrorFromError(err)
	}
	return deleted, nil
}

// Ping verifies connectivity to the back-end
func (p *PersonInterceptor) Ping(ctx context.Context) error {
	err := p.db.Ping(ctx)
	if err != nil {
		return br.NewDataAccessErrorFromError(err)
	}
	return nil
}
