package service

import (
	"context"
	"errors"

	br "github.com/morimar32/helpers/errors"
	"github.com/morimar32/helpers/proto"

	person "personsvc/generated"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewPersonService Factory to create a new PersonService instance
func NewPersonService(db *PersonRepository, Log *zap.Logger) person.PersonServer {
	ret := &PersonService{
		log:         Log,
		interceptor: NewPersonInterceptor(db),
	}
	return ret
}

// PersonService service for interacting with person records
type PersonService struct {
	person.PersonServer
	log         *zap.Logger
	interceptor PersonInterceptor
}

// GetPerson returns a person record from the system
func (s *PersonService) GetPerson(ctx context.Context, req *person.PersonRequest) (*person.PersonResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "must pass in a value")
	}
	model, err := s.interceptor.GetPerson(ctx, req.Id)
	if err != nil {
		return nil, translateError(err)
	}
	defer PutPersonEntity(model)

	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// GetPersons Returns all persons in the database, alphabetically
func (s *PersonService) GetPersons(ctx context.Context, in *empty.Empty) (*person.PersonListResponse, error) {
	list, err := s.interceptor.GetPersons(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	ret := personEntityArrayToPersonResponseArray(list)
	return ret, nil
}

// AddPerson Adds a person to the system
func (s *PersonService) AddPerson(ctx context.Context, req *person.AddPersonRequest) (*person.PersonResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "must pass in a value")
	}
	add := GetPersonEntity()
	defer PutPersonEntity(add)
	add.Bind("",
		req.FirstName,
		proto.StringValueToString(req.MiddleName),
		req.LastName,
		proto.StringValueToString(req.Suffix),
		nil, nil)

	model, err := s.interceptor.AddPerson(ctx, add)
	if err != nil {
		return nil, translateError(err)
	}
	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// UpdatePerson updates an existing person in the system
func (s *PersonService) UpdatePerson(ctx context.Context, req *person.UpdatePersonRequest) (*person.PersonResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "must pass in a value")
	}
	update := GetPersonEntity()
	defer PutPersonEntity(update)

	update.Bind(req.Id,
		req.FirstName,
		proto.StringValueToString(req.MiddleName),
		req.LastName,
		proto.StringValueToString(req.Suffix), nil, nil)

	model, err := s.interceptor.UpdatePerson(ctx, update)
	if err != nil {
		return nil, translateError(err)
	}
	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// DeletePerson Deletes a person from the system
func (s *PersonService) DeletePerson(ctx context.Context, req *person.PersonRequest) (*empty.Empty, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "must pass in a value")
	}
	_, err := s.interceptor.DeletePerson(ctx, req.Id)
	if err != nil {
		return nil, translateError(err)
	}
	return &empty.Empty{}, nil
}

// Ping checks connectivity on dependencies
func (s *PersonService) Ping(ctx context.Context, in *empty.Empty) (*empty.Empty, error) {
	if err := s.interceptor.Ping(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func personEntityToPersonResponse(entity *PersonEntity) *person.PersonResponse {
	resp := &person.PersonResponse{
		Id:         entity.ID,
		FirstName:  entity.firstName,
		MiddleName: proto.StringToStringValue(entity.middleName),
		LastName:   entity.lastName,
		Suffix:     proto.StringToStringValue(entity.suffix),
		Created:    proto.TimeToTimestamp(entity.created),
		Updated:    proto.TimeToTimestamp(entity.updated),
	}

	return resp
}

func personEntityArrayToPersonResponseArray(entities []*PersonEntity) *person.PersonListResponse {
	ret := &person.PersonListResponse{
		Persons: make([]*person.PersonResponse, len(entities)),
	}
	for _, item := range entities {
		val := personEntityToPersonResponse(item)
		ret.Persons = append(ret.Persons, val)
		PutPersonEntity(item)
	}
	return ret
}

func translateError(err error) error {
	if err != nil {
		if errors.Is(err, &br.DataAccessError{}) {
			return status.Errorf(codes.Internal, err.Error())
		}
		if errors.Is(err, &br.ValidationError{}) {
			return status.Errorf(codes.InvalidArgument, err.Error())
		}
		return status.Errorf(codes.Unknown, err.Error())
	}
	return nil
}
