package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"personsvc/internal"

	retry "github.com/morimar32/helpers/retry"

	"sync"
	"time"

	"github.com/google/uuid"
	dbhelper "github.com/morimar32/helpers/database"
)

type personDbResult struct {
	id         []byte
	firstname  string
	middlename sql.NullString
	lastname   string
	suffix     sql.NullString
	created    time.Time
	updated    sql.NullTime
}

var (
	getQueryResults = sync.Pool{
		New: func() interface{} {
			return &personDbResult{}
		},
	}
	dbOnce     sync.Once
	dbInstance PersonDB
)

// NewPersonDB Singleton surrounding a PersonDataAccess instance
func NewPersonDB(connection *sql.DB, pol *retry.DbRetry) *PersonDB {
	dbOnce.Do(func() {
		db := &PersonDB{
			connection: connection,
			getStmt:    getStatement(connection, getSQL),
			listStmt:   getStatement(connection, listSQL),
			updateStmt: getStatement(connection, updateSQL),
			deleteStmt: getStatement(connection, deleteSQL),
			addStmt:    getStatement(connection, addSQL),
			policy:     pol,
		}
		dbInstance = *db
		fmt.Println("Connection configured")
	})
	return &dbInstance
}

func getStatement(connection *sql.DB, query string) *sql.Stmt {
	stmt, err := connection.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

// PersonDB specific implementation for interacting with persons in the system
type PersonDB struct {
	connection *sql.DB
	getStmt    *sql.Stmt
	listStmt   *sql.Stmt
	updateStmt *sql.Stmt
	deleteStmt *sql.Stmt
	addStmt    *sql.Stmt
	policy     *retry.DbRetry
}

const (
	getSQL    = "SELECT TOP 1 Id, FirstName, MiddleName, LastName, Suffix, CreatedDateTime, UpdatedDateTime FROM UserName WITH (NOLOCK) WHERE Id = @Id"
	listSQL   = "SELECT TOP 10 Id, FirstName, MiddleName, LastName, Suffix, CreatedDateTime, UpdatedDateTime FROM UserName WITH (NOLOCK) ORDER BY LastName, FirstName"
	updateSQL = "UPDATE UserName SET FirstName = ?, MiddleName = ?, LastName = ?, Suffix = ?, UpdatedDateTime = CURRENT_TIMESTAMP WHERE Id = ?"
	deleteSQL = "DELETE FROM UserName WHERE Id = ?"
	addSQL    = `DECLARE @ID UNIQUEIDENTIFIER; SET @ID = NEWID(); 
						INSERT INTO UserName ( Id, FirstName, MiddleName, LastName, Suffix ) 
						VALUES( @ID, ?, ?, ?, ?); 
						SELECT TOP 1 Id, FirstName, MiddleName, LastName, Suffix, CreatedDateTime, UpdatedDateTime FROM UserName WITH (NOLOCK) WHERE Id = @ID`
)

// Ping verify connectivity to the database
func (db *PersonDB) Ping(ctx context.Context) error {
	err := db.connection.Ping()
	if err != nil {
		return nil
	}
	return err
}

// Get Returns the Person associated with the identifier
func (db *PersonDB) Get(ctx context.Context, id string) (*PersonEntity, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var entity *PersonEntity = nil
	var err error

	var (
		db_id         []byte
		db_firstname  string
		db_middlename sql.NullString
		db_lastname   string
		db_suffix     sql.NullString
		db_created    time.Time
		db_updated    sql.NullTime
	)

	if err = db.getStmt.QueryRowContext(ctx, sql.Named("Id", id)).Scan(&db_id, &db_firstname, &db_middlename, &db_lastname, &db_suffix, &db_created, &db_updated); err != nil {
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, errors.Join(fmt.Errorf("GetPerson - queryrow context"), err, internal.ErrValidation)
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entity = GetPersonEntity()
	entity.ID = dbhelper.GetGUIDString(db_id)
	entity.FirstName = db_firstname
	entity.MiddleName = db_middlename.String
	entity.LastName = db_lastname
	entity.Suffix = db_suffix.String
	entity.Created = &db_created
	entity.Updated = nil
	if db_updated.Valid {
		entity.Updated = &db_updated.Time
	}

	return entity, nil
}

func (db *PersonDB) Get_WithTx(ctx context.Context, tx *sql.Tx, id string) (*PersonEntity, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var entity *PersonEntity = nil
	tx_started := false
	var err error
	if tx == nil {
		tx, err = db.connection.BeginTx(ctx, &sql.TxOptions{})
		tx_started = true
		defer tx.Rollback()
		if err != nil {
			return nil, err
		}
	}
	var (
		db_id         []byte
		db_firstname  string
		db_middlename sql.NullString
		db_lastname   string
		db_suffix     sql.NullString
		db_created    time.Time
		db_updated    sql.NullTime
	)

	if err = db.policy.QueryRowContext(ctx, tx, db.getStmt, sql.Named("Id", id)).Scan(&db_id, &db_firstname, &db_middlename, &db_lastname, &db_suffix, &db_created, &db_updated); err != nil {
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, errors.Join(fmt.Errorf("GetPerson - queryrow context"), err, internal.ErrValidation)
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entity = GetPersonEntity()
	entity.ID = dbhelper.GetGUIDString(db_id)
	entity.FirstName = db_firstname
	entity.MiddleName = db_middlename.String
	entity.LastName = db_lastname
	entity.Suffix = db_suffix.String
	entity.Created = &db_created
	entity.Updated = nil
	if db_updated.Valid {
		entity.Updated = &db_updated.Time
	}

	if tx_started {
		tx.Commit()
	}
	return entity, nil
}

// GetList returns a list of Person entities from the system
func (db *PersonDB) GetList(ctx context.Context) ([]*PersonEntity, error) {
	ret := make([]*PersonEntity, 0)

	var results = getQueryResults.Get().(*personDbResult)
	defer getQueryResults.Put(results)

	total, err := dbhelper.QueryStatement(ctx, db.listStmt, func(rows *sql.Rows) error {
		if err := rows.Scan(&results.id, &results.firstname, &results.middlename, &results.lastname, &results.suffix, &results.created, &results.updated); err != nil {
			return err
		}
		item := GetPersonEntity()
		item.Bind(
			dbhelper.GetGUIDString(results.id),
			results.firstname,
			results.middlename.String,
			results.lastname,
			results.suffix.String,
			&results.created,
			dbhelper.NullTimeToTime(results.updated))
		ret = append(ret, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if total <= 0 {
		return nil, nil
	}
	return ret, nil
}

// Add creates a Person record in the system
func (db *PersonDB) Add(ctx context.Context, tx *sql.Tx, add *PersonEntity) (*PersonEntity, error) {

	tx_started := false
	var err error
	if tx == nil {
		tx, err = db.connection.BeginTx(ctx, &sql.TxOptions{})
		tx_started = true
		defer tx.Rollback()
		if err != nil {
			return nil, err
		}
	}
	var Id = uuid.NewString()
	stmt := tx.StmtContext(ctx, db.addStmt)
	_, err = stmt.ExecContext(ctx, Id, add.FirstName, add.MiddleName, add.LastName, add.Suffix)
	if err != nil {
		return nil, fmt.Errorf("PersonDB: AddPerson - Failed exec context: %w", err)
	}

	ret, err := db.Get_WithTx(ctx, tx, Id)
	if err != nil {
		return nil, fmt.Errorf("PersonDB: AddPerson - Failed querying added record: %w", err)
	}

	if tx_started {
		tx.Commit()
	}

	return ret, nil
}

// Update updates a person record in the system
func (db *PersonDB) Update(ctx context.Context, update *PersonEntity) (*PersonEntity, error) {

	total, err := dbhelper.ExecuteStatementNonQuery(ctx, db.updateStmt, update.FirstName, update.MiddleName, update.LastName, update.Suffix, update.ID)
	if err != nil {
		return nil, err
	}
	if total <= 0 {
		return nil, fmt.Errorf("No record found for %s", update.ID)
	}
	return db.Get_WithTx(ctx, nil, update.ID)
}

// Delete removes a Person from the system
func (db *PersonDB) Delete(ctx context.Context, id string) (bool, error) {
	total, err := dbhelper.ExecuteStatementNonQuery(ctx, db.deleteStmt, id)
	if err != nil {
		return false, err
	}
	if total <= 0 {
		return false, fmt.Errorf("No record found for %s", id)
	}
	return true, nil
}
