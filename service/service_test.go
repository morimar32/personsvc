package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"testing"
	"time"

	person "personsvc/generated"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
)

const (
	codeCoverageThreshold = 0.7
)

func getMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, _ := sqlmock.New()
	mock.ExpectPrepare("^SELECT (.+) FROM UserName (.+)")  //get
	mock.ExpectPrepare("^SELECT (.+) FROM UserName (.+)")  //list
	mock.ExpectPrepare("^UPDATE UserName (.+)")            //update
	mock.ExpectPrepare("^DELETE FROM UserName WHERE (.+)") //delete
	mock.ExpectPrepare("^(.+) INSERT INTO UserName (.+)")  //add
	return db, mock
}

func getParams(args ...interface{}) []driver.NamedValue {
	nvargs := make([]driver.NamedValue, len(args))
	for i := 0; i < len(args); i++ {
		nv := &driver.NamedValue{
			Ordinal: i + 1,
			Value:   args[i],
		}
		nvargs[i] = *nv
	}
	return nvargs
}
func TestMain(m *testing.M) {
	rc := m.Run()

	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < codeCoverageThreshold {
			fmt.Println("Tests passed but coverage failed at", c)
			rc = -1
		}
	}
	os.Exit(rc)
}

func TestGetPerson(t *testing.T) {
	ctx := context.Background()
	svc := &PersonService{}
	db, mock := getMock()
	defer db.Close()
	repo := prepNewRepository(db)
	interceptor := NewPersonInterceptor(repo)
	svc.interceptor = interceptor

	var testData = []struct {
		id       string
		expected bool
		msg      string
	}{
		{"", false, "empty request"},
		{"01234567-89ab-cdef-0123-456789abcdef", true, "happy path"},
	}
	_, err := svc.GetPerson(ctx, nil)
	if err == nil {
		t.Errorf("nil request should fail")
	}

	for _, tt := range testData {
		t.Run(tt.msg, func(t *testing.T) {

			req := &person.PersonRequest{
				Id: tt.id,
			}
			if tt.expected {
				expectedVals := sqlmock.NewRows([]string{"Id", "FirstName", "MiddleName", "LastName", "Suffix", "CreatedDateTime", "UpdatedDateTime"}).AddRow(tt.id, "first", "middle", "last", "suffix", "created", "updated")
				mock.ExpectQuery("^SELECT (.+)").RowsWillBeClosed().WillReturnRows(expectedVals)
			}
			_, err := svc.GetPerson(ctx, req)
			if err != nil && tt.expected {
				t.Errorf(err.Error())
			}
			if err == nil && !tt.expected {
				t.Errorf("expected failure did not happen")
			}
		})
	}
}

func TestGetPersons(t *testing.T) {
	ctx := context.Background()
	svc := &PersonService{}
	db, mock := getMock()
	defer db.Close()
	repo := prepNewRepository(db)
	interceptor := NewPersonInterceptor(repo)
	svc.interceptor = interceptor

	var testData = []struct {
		id       string
		expected bool
		msg      string
	}{
		{"", false, "empty request"},
		{"01234567-89ab-cdef-0123-456789abcdef", true, "happy path"},
	}
	_, err := svc.GetPersons(ctx, nil)
	if err == nil {
		t.Errorf("nil request should fail")
	}

	for _, tt := range testData {
		t.Run(tt.msg, func(t *testing.T) {
			now := time.Now()
			req := &empty.Empty{}
			if tt.expected {
				expectedVals := sqlmock.NewRows([]string{"Id", "FirstName", "MiddleName", "LastName", "Suffix", "CreatedDateTime", "UpdatedDateTime"}).AddRow(tt.id, "first", "middle", "last", "suffix", now, now)
				mock.ExpectQuery("^SELECT (.+)").RowsWillBeClosed().WillReturnRows(expectedVals)
			}
			_, err := svc.GetPersons(ctx, req)
			if err != nil && tt.expected {
				t.Errorf(err.Error())
			}
			if err == nil && !tt.expected {
				t.Errorf("expected failure did not happen")
			}
		})
	}
}

func TestAddPerson(t *testing.T) {
	ctx := context.Background()
	svc := &PersonService{}
	db, mock := getMock()
	defer db.Close()
	repo := prepNewRepository(db)
	interceptor := NewPersonInterceptor(repo)
	svc.interceptor = interceptor

	var testData = []struct {
		id                                      []byte
		firstname, middlename, lastname, suffix string
		expected                                bool
		msg                                     string
	}{
		{[]byte(""), "", "", "", "", false, "empty request"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "", "", "", "", false, "firstname required"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "", "", "", false, "lastname required"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy01234567890123456789012345678901234567890123456789", "", "McTesterFace", "", false, "firstname too long"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "Testy01234567890123456789012345678901234567890123456789", "McTesterFace", "", false, "middlename too long"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "", "Testy0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789", "", false, "lastname too long"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "", "McTesterFace", "Jr01234567890123456789", false, "suffix too long"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "", "McTesterFace", "Jr", true, "happy path"},
	}

	_, err := svc.AddPerson(ctx, nil)
	if err == nil {
		t.Errorf("nil request should fail")
	}
	for _, tt := range testData {

		t.Run(tt.msg, func(t *testing.T) {

			req := &person.AddPersonRequest{
				FirstName:  tt.firstname,
				MiddleName: &wrappers.StringValue{Value: tt.middlename},
				LastName:   tt.lastname,
				Suffix:     &wrappers.StringValue{Value: tt.suffix},
			}

			if tt.expected {
				now := time.Now()
				expectedVals := sqlmock.NewRows([]string{"Id", "FirstName", "MiddleName", "LastName", "Suffix", "CreatedDateTime", "UpdatedDateTime"}).AddRow(tt.id, tt.firstname, tt.middlename, tt.lastname, tt.suffix, now, now)

				mock.ExpectQuery("^(.+) INSERT INTO UserName (.+)").WithArgs(tt.firstname, tt.middlename, tt.lastname, tt.suffix).WillReturnRows(expectedVals)
			}

			_, err := svc.AddPerson(ctx, req)
			if err != nil && tt.expected {
				t.Errorf(err.Error())
			}
			if err == nil && !tt.expected {
				t.Errorf("expected failure did not happen")
			}

		})

	}
}

func TestUpdatePerson(t *testing.T) {
	ctx := context.Background()
	svc := &PersonService{}
	db, mock := getMock()
	defer db.Close()
	repo := prepNewRepository(db)
	interceptor := NewPersonInterceptor(repo)
	svc.interceptor = interceptor

	var testData = []struct {
		id                                      []byte
		firstname, middlename, lastname, suffix string
		expected                                bool
		msg                                     string
	}{
		{[]byte(""), "", "", "", "", false, "empty request"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "", "", "", "", false, "firstname required"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "", "", "", false, "lastname required"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy01234567890123456789012345678901234567890123456789", "", "McTesterFace", "", false, "firstname too long"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "Testy01234567890123456789012345678901234567890123456789", "McTesterFace", "", false, "middlename too long"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "", "Testy0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789", "", false, "lastname too long"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "", "McTesterFace", "Jr01234567890123456789", false, "suffix too long"},
		{[]byte("fffe7a94-64fa-4ac5-9ae4-19387b66582d"), "Testy", "", "McTesterFace", "Jr", true, "happy path"},
	}

	_, err := svc.UpdatePerson(ctx, nil)
	if err == nil {
		t.Errorf("nil request should fail")
	}
	for _, tt := range testData {

		t.Run(tt.msg, func(t *testing.T) {

			req := &person.UpdatePersonRequest{
				Id:         string(tt.id),
				FirstName:  tt.firstname,
				MiddleName: &wrappers.StringValue{Value: tt.middlename},
				LastName:   tt.lastname,
				Suffix:     &wrappers.StringValue{Value: tt.suffix},
			}

			if tt.expected {
				expectedVals := sqlmock.NewRows([]string{"Id", "FirstName", "MiddleName", "LastName", "Suffix"}).AddRow(tt.id, tt.firstname, tt.middlename, tt.lastname, tt.suffix)
				updateResult := sqlmock.NewResult(123, 123)
				mock.ExpectExec("^UPDATE UserName (.+)").WithArgs(tt.firstname, tt.middlename, tt.lastname, tt.suffix, "fffe7a94-64fa-4ac5-9ae4-19387b66582d").WillReturnResult(updateResult)
				mock.ExpectQuery("^SELECT (.+)").WillReturnRows(expectedVals)
			}

			_, err := svc.UpdatePerson(ctx, req)
			if err != nil && tt.expected {
				t.Errorf(err.Error())
			}
			if err == nil && !tt.expected {
				t.Errorf("expected failure did not happen")
			}

		})

	}
}

func TestDeletePerson(t *testing.T) {
	ctx := context.Background()
	svc := &PersonService{}
	db, mock := getMock()
	defer db.Close()
	repo := prepNewRepository(db)
	interceptor := NewPersonInterceptor(repo)
	svc.interceptor = interceptor

	var testData = []struct {
		id       string
		expected bool
		msg      string
	}{
		{"", false, "empty request"},
		{"01234567-89ab-cdef-0123-456789abcdef", true, "happy path"},
	}
	_, err := svc.DeletePerson(ctx, nil)
	if err == nil {
		t.Errorf("nil request should fail")
	}

	for _, tt := range testData {
		t.Run(tt.msg, func(t *testing.T) {

			req := &person.PersonRequest{
				Id: tt.id,
			}
			if tt.expected {

				updateResult := sqlmock.NewResult(123, 123)
				mock.ExpectExec("^DELETE FROM UserName WHERE (.+)").WithArgs(tt.id).WillReturnResult(updateResult)
			}
			_, err := svc.DeletePerson(ctx, req)
			if err != nil && tt.expected {
				t.Errorf(err.Error())
			}
			if err == nil && !tt.expected {
				t.Errorf("expected failure did not happen")
			}
		})
	}
}

func TestPing(t *testing.T) {
	ctx := context.Background()
	svc := &PersonService{}
	db, mock := getMock()
	defer db.Close()
	repo := prepNewRepository(db)
	interceptor := NewPersonInterceptor(repo)
	svc.interceptor = interceptor

	mock.ExpectPing()
	_, err := svc.Ping(ctx, nil)
	if err != nil {
		t.Errorf("nil request should not fail")
	}

}

func TestNewPersonService(t *testing.T) {
	db, _ := getMock()
	defer db.Close()
	repo := prepNewRepository(db)

	ret := NewPersonService(repo, nil)
	if ret == nil {
		t.Errorf("basic NewPersonService call should not fail")
	}
}
