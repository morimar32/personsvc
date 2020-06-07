package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	person "personsvc/generated"

	"github.com/golang/protobuf/ptypes/wrappers"
)

const (
	codeCoverageThreshold = 0.2
)

type MockPersonRepository struct {
	getShouldFail    bool
	updateShouldFail bool
}

func getMock() IPersonRepository {
	db := &MockPersonRepository{}
	return db
}

func (m *MockPersonRepository) GetList(ctx context.Context) ([]*PersonEntity, error) {
	if m.getShouldFail {
		return nil, fmt.Errorf("Could not get list")
	}
	return nil, nil
}

func (m *MockPersonRepository) Add(ctx context.Context, add *PersonEntity) (*PersonEntity, error) {
	if m.getShouldFail {
		return nil, fmt.Errorf("Could not add")
	}
	return &PersonEntity{}, nil
}
func (m *MockPersonRepository) Delete(ctx context.Context, id string) (bool, error) {
	if m.getShouldFail {
		return false, fmt.Errorf("Count not delete")
	}
	return true, nil
}

func (m *MockPersonRepository) Ping(ctx context.Context) error {
	if m.getShouldFail {
		return fmt.Errorf("Could not connect")
	}
	return nil
}

func (m *MockPersonRepository) Get(ctx context.Context, id string) (*PersonEntity, error) {
	if m.getShouldFail {
		return nil, fmt.Errorf("Could not query")
	}
	val := &PersonEntity{
		ID:        "01234567-89ab-cdef-0123-456789abcdef",
		firstName: "Testy",
		lastName:  "McTesterFace",
	}
	return val, nil
}

func (m *MockPersonRepository) Update(ctx context.Context, update *PersonEntity) (*PersonEntity, error) {
	if m.updateShouldFail {
		return nil, fmt.Errorf("Could not save")
	}
	val := &PersonEntity{
		ID:        "01234567-89ab-cdef-0123-456789abcdef",
		firstName: "Testy",
		lastName:  "McTesterFace",
	}
	return val, nil
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
	svc.db = getMock()

	var testData = []struct {
		id       string
		expected bool
		msg      string
	}{
		{"", false, "empty request"},
		{"123", true, "happy path"},
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

func TestUpdatePerson(t *testing.T) {
	ctx := context.Background()
	svc := &PersonService{}
	svc.db = getMock()

	var testData = []struct {
		id, firstname, middlename, lastname, suffix string
		expected                                    bool
		msg                                         string
	}{
		{"", "", "", "", "", false, "empty request"},
		{"123", "", "", "", "", false, "firstname required"},
		{"123", "Testy", "", "", "", false, "lastname required"},
		{"123", "Testy01234567890123456789012345678901234567890123456789", "", "McTesterFace", "", false, "firstname too long"},
		{"123", "Testy", "Testy01234567890123456789012345678901234567890123456789", "McTesterFace", "", false, "middlename too long"},
		{"123", "Testy", "", "Testy0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789", "", false, "lastname too long"},
		{"123", "Testy", "", "McTesterFace", "Jr01234567890123456789", false, "suffix too long"},
		{"123", "Testy", "", "McTesterFace", "Jr", true, "happy path"},
	}

	_, err := svc.UpdatePerson(ctx, nil)
	if err == nil {
		t.Errorf("nil request should fail")
	}
	for _, tt := range testData {

		t.Run(tt.msg, func(t *testing.T) {

			req := &person.UpdatePersonRequest{
				Id:         tt.id,
				FirstName:  tt.firstname,
				MiddleName: &wrappers.StringValue{Value: tt.middlename},
				LastName:   tt.lastname,
				Suffix:     &wrappers.StringValue{Value: tt.suffix},
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
