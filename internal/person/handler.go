package service

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	br "github.com/morimar32/helpers/errors"
)

type PersonHandler struct {
	Db PersonDB
}

func NewPersonHandler(repo *PersonDB) PersonHandler {
	val := &PersonHandler{
		Db: *repo,
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
func (p *PersonHandler) GetPerson(ctx context.Context, id string) (*PersonEntity, error) {
	if len(id) <= 0 {
		return nil, br.NewValidationError("Id is required")
	}

	var result = make(chan *PersonEntity)
	var e = make(chan error)
	go func(ctx context.Context, id string, result chan<- *PersonEntity, e chan<- error) {
		tx, err := p.Db.connection.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			e <- err
		} else {
			defer tx.Rollback()
			model, err := p.Db.Get(ctx, tx, id)
			if err != nil {
				e <- br.NewDataAccessErrorFromError(err)
			} else if model == nil {
				e <- br.NewValidationError("Person not found")
			}
			tx.Commit()
			result <- model
		}
	}(ctx, id, result, e)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case model := <-result:
		return model, nil
	case err := <-e:
		return nil, err
	}
}

// GetPersons returns a list of people
func (p *PersonHandler) GetPersons(ctx context.Context) ([]*PersonEntity, error) {
	vals, err := p.Db.GetList(ctx)
	if err != nil {
		return nil, br.NewDataAccessErrorFromError(err)
	}
	if vals == nil {
		return nil, br.NewValidationError("No Persons found")
	}
	return vals, nil
}

// AddPerson creates a new person record in the system
func (p *PersonHandler) AddPerson(ctx context.Context, add *PersonEntity) (*PersonEntity, error) {
	err := p.validateCommon(add)
	if err != nil {
		return nil, err
	}

	person, err := p.Db.Add(ctx, add)
	if err != nil {
		return nil, br.NewDataAccessErrorFromError(err)
	}
	return person, nil
}

// UpdatePerson updates an existing person record in the system
func (p *PersonHandler) UpdatePerson(ctx context.Context, update *PersonEntity) (*PersonEntity, error) {
	err := p.validateUpdate(update)
	if err != nil {
		return nil, err
	}
	person, err := p.Db.Update(ctx, update)
	if err != nil {
		return nil, br.NewDataAccessErrorFromError(err)
	}
	return person, nil
}

// DeletePerson removes a person from the system
func (p *PersonHandler) DeletePerson(ctx context.Context, id string) (bool, error) {
	if len(id) < 36 {
		return false, br.NewValidationError("Invalid id")
	}
	deleted, err := p.Db.Delete(ctx, id)
	if err != nil {
		return false, br.NewDataAccessErrorFromError(err)
	}
	return deleted, nil
}

// Ping verifies connectivity to the back-end
func (p *PersonHandler) Ping(ctx context.Context) error {
	err := p.Db.Ping(ctx)
	if err != nil {
		return br.NewDataAccessErrorFromError(err)
	}
	return nil
}
