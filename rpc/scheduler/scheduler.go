package schedulersvc

import (
	"golang.org/x/net/context"

	"github.com/iampigeon/pigeon"
	pb "github.com/iampigeon/pigeon/proto"
	"github.com/iampigeon/pigeon/scheduler"
	"github.com/oklog/ulid"
)

var _ pb.SchedulerServiceServer = (*Service)(nil)

// Service ...
type Service struct {
	schedulerSvc pigeon.SchedulerService
}

// New ...
func New(config scheduler.StorageConfig) (*Service, error) {
	sSvc, err := scheduler.NewStoreBackend(config)
	if err != nil {
		return nil, err
	}
	return &Service{
		schedulerSvc: sSvc,
	}, nil
}

// Put ...
func (s *Service) Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error) {
	id, err := ulid.Parse(r.Id)
	if err != nil {
		return nil, err
	}

	if err := s.schedulerSvc.Put(id, r.Content, pigeon.NetAddr(r.Endpoint), pigeon.StatusPending, r.SubjectId); err != nil {
		return nil, err
	}

	return &pb.PutResponse{}, nil
}

func (s *Service) DeleteByID(ctx context.Context, r *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	_, err := ulid.Parse(r.Id)
	if err != nil { // (Coke, read this): We only need to check if is a valid ulid
		return nil, err
	}

	if err := s.schedulerSvc.DeleteByID(id); err != nil {
		return nil, err
	}

	return &pb.DeleteResponse{}, nil
}

// Get ...
func (s *Service) Get(ctx context.Context, r *pb.GetRequest) (*pb.GetResponse, error) {
	id, err := ulid.Parse(r.Id)
	if err != nil {
		return nil, err
	}

	msg, err := s.schedulerSvc.Get(id)
	if err != nil {
		return nil, err
	}

	return &pb.GetResponse{
		Message: &pb.Message{
			Id:       r.Id,
			Content:  msg.Content,
			Endpoint: string(msg.Endpoint),
			Status:   string(msg.Status),
		},
	}, nil
}

// Update ...
func (s *Service) Update(ctx context.Context, r *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	id, err := ulid.Parse(r.Id)
	if err != nil {
		return nil, err
	}

	if err := s.schedulerSvc.Update(id, r.Content); err != nil {
		return nil, err
	}

	return &pb.UpdateResponse{}, nil
}
