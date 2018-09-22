package schedulersvc

import (
	"golang.org/x/net/context"

	"github.com/iampigeon/pigeon"
	pb "github.com/iampigeon/pigeon/proto"
	"github.com/iampigeon/pigeon/scheduler"
	"github.com/oklog/ulid"
)

var _ pb.SchedulerServiceServer = (*Service)(nil)

type Service struct {
	s pigeon.SchedulerService
}

func New(config scheduler.StorageConfig) *Service {
	return &Service{
		s: scheduler.New(config),
	}
}

func (s *Service) Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error) {
	id, err := ulid.Parse(r.Id)
	if err != nil {
		return nil, err
	}

	if err := s.s.Put(id, r.Content, pigeon.NetAddr(r.Endpoint)); err != nil {
		return nil, err
	}

	return &pb.PutResponse{}, nil
}
func (s *Service) Get(ctx context.Context, r *pb.GetRequest) (*pb.GetResponse, error) {
	id, err := ulid.Parse(r.Id)
	if err != nil {
		return nil, err
	}

	msg, err := s.s.Get(id)
	if err != nil {
		return nil, err
	}

	return &pb.GetResponse{
		Message: &pb.Message{
			Id:       r.Id,
			Content:  msg.Content,
			Endpoint: string(msg.Endpoint),
		},
	}, nil
}
func (s *Service) Update(ctx context.Context, r *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	id, err := ulid.Parse(r.Id)
	if err != nil {
		return nil, err
	}

	if err := s.s.Update(id, r.Content); err != nil {
		return nil, err
	}

	return &pb.UpdateResponse{}, nil
}
