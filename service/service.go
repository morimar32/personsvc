package service

import (
	"context"

	"github.com/morimar32/helpers/proto"

	person "personsvc/generated"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewPersonService Factory to create a new PersonService instance
func NewPersonService(db *IPersonRepository, Log *zap.Logger) person.PersonServer {
	ret := &PersonService{}
	ret.db = *db
	ret.log = Log
	return ret
}

// PersonService service for interacting with person records
type PersonService struct {
	db IPersonRepository
	person.PersonServer
	log *zap.Logger
}

/*
// AuthFuncOverride is called instead of exampleAuthFunc
func (g *PersonService) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	log.Println("client is calling method:", fullMethodName)
	return ctx, nil
}
*/
// GetPerson returns a person record from the system
func (s *PersonService) GetPerson(ctx context.Context, req *person.PersonRequest) (*person.PersonResponse, error) {
	if err := s.validateGet(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	model, err := s.db.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if model == nil {
		return nil, status.Errorf(codes.NotFound, "Person not found")
	}
	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// GetPersons Returns all persons in the database, alphabetically
func (s *PersonService) GetPersons(ctx context.Context, in *empty.Empty) (*person.PersonListResponse, error) {
	list, err := s.db.GetList(ctx)
	if err != nil {
		return nil, err
	}
	ret := &person.PersonListResponse{
		Persons: make([]*person.PersonResponse, 0),
	}
	for _, item := range list {
		val := personEntityToPersonResponse(item)
		ret.Persons = append(ret.Persons, val)
	}
	return ret, nil
}

// AddPerson Adds a person to the system
func (s *PersonService) AddPerson(ctx context.Context, req *person.AddPersonRequest) (*person.PersonResponse, error) {
	add := &PersonEntity{
		firstName:  req.FirstName,
		middleName: proto.StringValueToString(req.MiddleName),
		lastName:   req.LastName,
		suffix:     proto.StringValueToString(req.Suffix),
	}
	model, err := s.db.Add(ctx, add)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// UpdatePerson updates an existing person in the system
func (s *PersonService) UpdatePerson(ctx context.Context, req *person.UpdatePersonRequest) (*person.PersonResponse, error) {
	if err := s.validateUpdate(req); err != nil {
		return nil, err
	}
	update := &PersonEntity{
		ID:         req.Id,
		firstName:  req.FirstName,
		middleName: req.MiddleName.Value,
		lastName:   req.LastName,
		suffix:     req.Suffix.Value,
	}

	model, err := s.db.Update(ctx, update)
	if err != nil {
		return nil, err
	}
	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// DeletePerson Deletes a person from the system
func (s *PersonService) DeletePerson(ctx context.Context, req *person.PersonRequest) (*empty.Empty, error) {
	if err := s.validateGet(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	_, err := s.db.Delete(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

// Ping checks connectivity on dependencies
func (s *PersonService) Ping(ctx context.Context, in *empty.Empty) (*empty.Empty, error) {
	if err := s.db.Ping(ctx); err != nil {
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
