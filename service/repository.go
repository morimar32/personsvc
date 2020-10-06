package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb" //mssql implementation
	dbhelper "github.com/morimar32/helpers/database"
)

type personDTO struct {
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
			return &personDTO{}
		},
	}
	dbOnce     sync.Once
	dbInstance *PersonRepository
)

// NewPersonRepository Singleton surrounding a PersonDataAccess instance
func NewPersonRepository(constring string) *PersonRepository {
	dbOnce.Do(func() {
		var err error
		c, err := dbhelper.InitConnection(constring, 50, 50, (1 * time.Hour))
		if err != nil {
			log.Fatal(err)
		}
		dbInstance = prepNewRepository(c)

		fmt.Println("Connection configured")
	})
	return dbInstance
}

func prepNewRepository(c *sql.DB) *PersonRepository {
	i := &PersonRepository{
		connection: c,
		getStmt:    getStatement(c, getSQL),
		listStmt:   getStatement(c, listSQL),
		updateStmt: getStatement(c, updateSQL),
		deleteStmt: getStatement(c, deleteSQL),
		addStmt:    getStatement(c, addSQL),
	}
	return i
}

func getStatement(connection *sql.DB, query string) *sql.Stmt {
	stmt, err := connection.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

// PersonRepository specific implementation for interacting with persons in the system
type PersonRepository struct {
	connection *sql.DB
	getStmt    *sql.Stmt
	listStmt   *sql.Stmt
	updateStmt *sql.Stmt
	deleteStmt *sql.Stmt
	addStmt    *sql.Stmt
}

const (
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
	err := db.connection.Ping()
	if err != nil {
		return nil
	}
	return err
}

// Get Returns the Person associated with the identifier
func (db *PersonRepository) Get(ctx context.Context, id string) (*PersonEntity, error) {
	var results = getQueryResults.Get().(*personDTO)
	defer getQueryResults.Put(results)

	var entity *PersonEntity = nil

	dbhelper.QueryStatement(ctx, db.getStmt, func(rows *sql.Rows) error {
		if err := rows.Scan(&results.id, &results.firstname, &results.middlename, &results.lastname, &results.suffix, &results.created, &results.updated); err != nil {
			return err
		}
		return nil
	}, id)

	entity = GetPersonEntity()
	entity = entity.Bind(
		GetGUIDString(results.id),
		results.firstname,
		results.middlename.String,
		results.lastname,
		results.suffix.String,
		&results.created,
		dbhelper.NullTimeToTime(results.updated))

	return entity, nil
}

// GetList returns a list of Person entities from the system
func (db *PersonRepository) GetList(ctx context.Context) ([]*PersonEntity, error) {
	ret := make([]*PersonEntity, 0)

	var results = getQueryResults.Get().(*personDTO)
	defer getQueryResults.Put(results)

	total, err := dbhelper.QueryStatement(ctx, db.listStmt, func(rows *sql.Rows) error {
		if err := rows.Scan(&results.id, &results.firstname, &results.middlename, &results.lastname, &results.suffix, &results.created, &results.updated); err != nil {
			return err
		}
		item := GetPersonEntity()
		item.Bind(
			GetGUIDString(results.id),
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
func (db *PersonRepository) Add(ctx context.Context, add *PersonEntity) (*PersonEntity, error) {
	var ret *PersonEntity = nil
	var results = getQueryResults.Get().(*personDTO)
	defer getQueryResults.Put(results)

	total, err := dbhelper.QueryStatement(ctx, db.addStmt, func(rows *sql.Rows) error {
		if err := rows.Scan(&results.id, &results.firstname, &results.middlename, &results.lastname, &results.suffix, &results.created, &results.updated); err != nil {
			return err
		}
		ret = GetPersonEntity()
		ret.Bind(
			GetGUIDString(results.id),
			results.firstname,
			results.middlename.String,
			results.lastname,
			results.suffix.String,
			&results.created,
			dbhelper.NullTimeToTime(results.updated))
		return nil
	}, add.firstName, add.middleName, add.lastName, add.suffix)
	if err != nil {
		return nil, err
	}
	if total <= 0 {
		return nil, nil
	}
	return ret, nil

}

// Update updates a person record in the system
func (db *PersonRepository) Update(ctx context.Context, update *PersonEntity) (*PersonEntity, error) {

	total, err := dbhelper.ExecuteStatementNonQuery(ctx, db.updateStmt, update.firstName, update.middleName, update.lastName, update.suffix, update.ID)
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
	total, err := dbhelper.ExecuteStatementNonQuery(ctx, db.deleteStmt, id)
	if err != nil {
		return false, err
	}
	if total <= 0 {
		return false, fmt.Errorf("No record found for %s", id)
	}
	return true, nil
}

func GetGUIDString(b []byte) string {
	if len(b) < 8 {
		return string(b)
	}
	b[0], b[1], b[2], b[3] = b[3], b[2], b[1], b[0]
	b[4], b[5] = b[5], b[4]
	b[6], b[7] = b[7], b[6]
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
