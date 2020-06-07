package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/morimar32/helpers/database"
)

// NewPersonRepository factory to create a New PersonDataAccess instance
func NewPersonRepository(constring string) IPersonRepository {
	db := &PersonRepository{}
	db.helper = &database.DbHelper{
		ConnectionString: constring,
	}
	return db
}

// IPersonRepository defines the database interactions for a person
type IPersonRepository interface {
	Ping(ctx context.Context) error
	Get(ctx context.Context, id string) (*PersonEntity, error)
	GetList(ctx context.Context) ([]*PersonEntity, error)
	Add(ctx context.Context, add *PersonEntity) (*PersonEntity, error)
	Update(ctx context.Context, update *PersonEntity) (*PersonEntity, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// PersonRepository specific implementation for interacting with persons in the system
type PersonRepository struct {
	helper *database.DbHelper
}

const (
	pingSQL   = "SELECT TOP 0 Id FROM UserName WITH (NOLOCK)"
	getSQL    = "SELECT TOP 1 Id, FirstName, MiddleName, LastName, Suffix, CreatedDateTime, UpdatedDateTime FROM UserName WITH (NOLOCK) WHERE Id = ?"
	listSQL   = "SELECT Id, FirstName, MiddleName, LastName, Suffix, CreatedDateTime, UpdatedDateTime FROM UserName WITH (NOLOCK) ORDER BY LastName, FirstName"
	updateSQL = "UPDATE UserName SET FirstName = ?, MiddleName = ?, LastName = ?, Suffix = ?, UpdatedDateTime = CURRENT_TIMESTAMP WHERE Id = ?"
	deleteSQL = "DELETE FROM UserName WHERE Id = ?"
	addSQL    = `DECLARE @ID UNIQUEIDENTIFIER; SET @ID = NEWID(); 
						INSERT INTO UserName ( Id, FirstName, MiddleName, LastName, Suffix ) 
						VALUES( @ID, ?, ?, ?, ?); 
						SELECT TOP 1 Id, FirstName, MiddleName, LastName, Suffix, CreatedDateTime, UpdatedDateTime FROM UserName WITH (NOLOCK) WHERE Id = @ID`
)

// Ping verify connectivity to the database
func (db *PersonRepository) Ping(ctx context.Context) error {
	err := db.helper.Query(ctx, pingSQL, func(*sql.Rows) error {
		return nil
	})
	return err
}

// Get Retrieves a person from the system
func (db *PersonRepository) Get(ctx context.Context, id string) (*PersonEntity, error) {
	var ret *PersonEntity
	var results = struct {
		id         []byte
		firstname  string
		middlename sql.NullString
		lastname   string
		suffix     sql.NullString
		created    time.Time
		updated    sql.NullTime
	}{}

	err := db.helper.Query(ctx, getSQL, func(rows *sql.Rows) error {
		if err := rows.Scan(&results.id, &results.firstname, &results.middlename, &results.lastname, &results.suffix, &results.created, &results.updated); err != nil {
			return err
		}
		ret = &PersonEntity{
			ID:         database.GetGUIDString(results.id),
			firstName:  results.firstname,
			middleName: results.middlename.String,
			lastName:   results.lastname,
			suffix:     results.suffix.String,
			created:    &results.created,
			updated:    &results.updated.Time,
		}
		return nil
	}, id)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// GetList returns a list of Person entities from the system
func (db *PersonRepository) GetList(ctx context.Context) ([]*PersonEntity, error) {
	ret := make([]*PersonEntity, 0)
	var results = struct {
		id         []byte
		firstname  string
		middlename sql.NullString
		lastname   string
		suffix     sql.NullString
		created    time.Time
		updated    sql.NullTime
	}{}
	err := db.helper.Query(ctx, listSQL, func(rows *sql.Rows) error {
		if err := rows.Scan(&results.id, &results.firstname, &results.middlename, &results.lastname, &results.suffix, &results.created, &results.updated); err != nil {
			return err
		}
		item := &PersonEntity{
			ID:         database.GetGUIDString(results.id),
			firstName:  results.firstname,
			middleName: results.middlename.String,
			lastName:   results.lastname,
			suffix:     results.suffix.String,
			created:    &results.created,
			updated:    &results.updated.Time,
		}
		ret = append(ret, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Add creates a Person record in the system
func (db *PersonRepository) Add(ctx context.Context, add *PersonEntity) (*PersonEntity, error) {
	var ret *PersonEntity
	var results = struct {
		id         []byte
		firstname  string
		middlename sql.NullString
		lastname   string
		suffix     sql.NullString
		created    time.Time
		updated    sql.NullTime
	}{}

	err := db.helper.Query(ctx, addSQL, func(rows *sql.Rows) error {
		if err := rows.Scan(&results.id, &results.firstname, &results.middlename, &results.lastname, &results.suffix, &results.created, &results.updated); err != nil {
			return err
		}
		ret = &PersonEntity{
			ID:         database.GetGUIDString(results.id),
			firstName:  results.firstname,
			middleName: results.middlename.String,
			lastName:   results.lastname,
			suffix:     results.suffix.String,
			created:    &results.created,
			updated:    &results.updated.Time,
		}
		return nil
	}, add.firstName, add.middleName, add.lastName, add.suffix)
	if err != nil {
		return nil, err
	}
	return ret, nil

}

// Update updates a person record in the system
func (db *PersonRepository) Update(ctx context.Context, update *PersonEntity) (*PersonEntity, error) {
	total, err := db.helper.ExecuteNonQuery(ctx, updateSQL, update.firstName, update.middleName, update.lastName, update.suffix, update.ID)
	if err != nil {
		return nil, err
	}
	if total <= 0 {
		return nil, fmt.Errorf("No record found for %s", update.ID)
	}
	return db.Get(ctx, update.ID)
}

// Delete removes a Person from the system
func (db *PersonRepository) Delete(ctx context.Context, id string) (bool, error) {
	total, err := db.helper.ExecuteNonQuery(ctx, deleteSQL, id)
	if err != nil {
		return false, err
	}
	if total <= 0 {
		return false, fmt.Errorf("No record found for %s", id)
	}
	return true, nil
}
