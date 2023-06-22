package service

import (
	"context"
	"database/sql"
	"errors"
	"personsvc/internal"

	outbox "github.com/morimar32/helpers/outbox"

	"sync"
	"sync/atomic"

	br "github.com/morimar32/helpers/errors"
)

const (
	PublishTopic       = "PersonTopic"
	AddPersonEventName = "PersonAdded"
)

type PersonHandler struct {
	Db     PersonDB
	outbox outbox.Outboxer
}

func NewPersonHandler(repo *PersonDB, outbox outbox.Outboxer) PersonHandler {
	val := &PersonHandler{
		Db:     *repo,
		outbox: outbox,
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
		return nil, errors.Join(errors.New("Person: GetPerson - Id is required"), internal.ErrValidation)
	}

	var result = make(chan *PersonEntity)
	var e = make(chan error)
	go func(ctx context.Context, id string, result chan<- *PersonEntity, e chan<- error) {
		tx, err := p.Db.connection.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			e <- errors.Join(errors.New("Person: GetPerson - Transaction"), err, internal.ErrSql)
			return
		}
		defer tx.Rollback()
		model, err := p.Db.Get(ctx, tx, id)
		if err != nil {
			e <- errors.Join(errors.New("Person: GetPerson - Get"), err, internal.ErrSql)
			return
		}
		if model == nil {
			e <- errors.Join(errors.New("Person: GetPerson - Person not found"), err, internal.ErrValidation)
			return
		}
		err = p.outbox.Publish(ctx, tx, "Person", "PersonRead", model)
		if err != nil {
			e <- errors.Join(errors.New("Person: GetPerson - Outbox"), err, internal.ErrSql)
			return
		}
		tx.Commit()
		result <- model
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

	var result = make(chan *PersonEntity)
	var e = make(chan error)
	go func(ctx context.Context, add *PersonEntity, result chan<- *PersonEntity, e chan<- error) {
		tx, err := p.Db.connection.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			e <- err
			return
		}
		defer tx.Rollback()
		val, err := p.Db.Add(ctx, tx, add)
		if err != nil {
			e <- err
			return
		}
		err = p.outbox.Publish(ctx, tx, PublishTopic, AddPersonEventName, val)
		if err != nil {
			e <- err
			return
		}

		tx.Commit()
		result <- val

	}(ctx, add, result, e)

	var val *PersonEntity = nil
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case val = <-result:
		break
	case err := <-e:
		return nil, br.NewDataAccessErrorFromError(err)
	}

	return val, nil
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
