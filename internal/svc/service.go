package service

import (
	"context"
	"errors"
	person "personsvc/internal/person"
	outbox "personsvc/pkg/outbox"
	"time"

	br "github.com/morimar32/helpers/errors"
	"github.com/morimar32/helpers/proto"

	pb "personsvc/generated"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewPersonService Factory to create a new PersonService instance
func NewPersonService(db *person.PersonDB, outbox outbox.Outboxer, Log *zap.Logger) pb.PersonServer {
	ret := &PersonService{
		log:     Log,
		handler: person.NewPersonHandler(db, outbox),
	}
	return ret
}

// PersonService service for interacting with person records
type PersonService struct {
	pb.PersonServer
	log     *zap.Logger
	handler person.PersonHandler
}

// GetPerson returns a person record from the system
func (s *PersonService) GetPerson(ctx context.Context, req *pb.PersonRequest) (*pb.PersonResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "must pass in a value")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(300 * time.Millisecond)
	}
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	model, err := s.handler.GetPerson(ctx, req.Id)
	if err != nil {
		return nil, translateError(err)
	}
	defer person.PutPersonEntity(model)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// GetPersons Returns all persons in the database, alphabetically
func (s *PersonService) GetPersons(ctx context.Context, in *empty.Empty) (*pb.PersonListResponse, error) {
	list, err := s.handler.GetPersons(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	ret := personEntityArrayToPersonResponseArray(list)
	return ret, nil
}

// AddPerson Adds a person to the system
func (s *PersonService) AddPerson(ctx context.Context, req *pb.AddPersonRequest) (*pb.PersonResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "must pass in a value")
	}
	add := person.GetPersonEntity()
	defer person.PutPersonEntity(add)
	add.Bind("",
		req.FirstName,
		proto.StringValueToString(req.MiddleName),
		req.LastName,
		proto.StringValueToString(req.Suffix),
		nil, nil)

	model, err := s.handler.AddPerson(ctx, add)
	if err != nil {
		return nil, translateError(err)
	}
	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// UpdatePerson updates an existing person in the system
func (s *PersonService) UpdatePerson(ctx context.Context, req *pb.UpdatePersonRequest) (*pb.PersonResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "must pass in a value")
	}
	update := person.GetPersonEntity()
	defer person.PutPersonEntity(update)

	update.Bind(req.Id,
		req.FirstName,
		proto.StringValueToString(req.MiddleName),
		req.LastName,
		proto.StringValueToString(req.Suffix), nil, nil)

	model, err := s.handler.UpdatePerson(ctx, update)
	if err != nil {
		return nil, translateError(err)
	}
	resp := personEntityToPersonResponse(model)
	return resp, nil
}

// DeletePerson Deletes a person from the system
func (s *PersonService) DeletePerson(ctx context.Context, req *pb.PersonRequest) (*empty.Empty, error) {
	_, err := s.handler.DeletePerson(ctx, req.Id)
	if err != nil {
		return nil, translateError(err)
	}
	return &empty.Empty{}, nil
}

// Ping checks connectivity on dependencies
func (s *PersonService) Ping(ctx context.Context, in *empty.Empty) (*empty.Empty, error) {
	if err := s.handler.Ping(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func personEntityToPersonResponse(entity *person.PersonEntity) *pb.PersonResponse {
	resp := &pb.PersonResponse{
		Id:         entity.ID,
		FirstName:  entity.FirstName,
		MiddleName: proto.StringToStringValue(entity.MiddleName),
		LastName:   entity.LastName,
		Suffix:     proto.StringToStringValue(entity.Suffix),
		Created:    proto.TimeToTimestamp(entity.Created),
		Updated:    proto.TimeToTimestamp(entity.Updated),
	}
	person.PutPersonEntity(entity)
	return resp
}

func personEntityArrayToPersonResponseArray(entities []*person.PersonEntity) *pb.PersonListResponse {
	ret := &pb.PersonListResponse{
		Persons: make([]*pb.PersonResponse, len(entities)),
	}
	for i, item := range entities {
		val := personEntityToPersonResponse(item)
		ret.Persons[i] = val
		person.PutPersonEntity(item)
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
