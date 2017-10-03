package backend

import (
	"context"
	"log"
	"net"

	"github.com/WiseGrowth/pigeon/pigeon"
	"github.com/WiseGrowth/pigeon/pigeon/proto"
	"google.golang.org/grpc"
)

func ListenAndServe(address string, backend pigeon.Backend) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	proto.RegisterBackendServiceServer(s, service{backend})

	return s.Serve(lis)
}

type service struct {
	backend pigeon.Backend
}

func (s *service) Aprove(ctx context.Context, r *proto.AproveRequest) (*proto.AproveResponse, error) {
	var (
		resp proto.AproveResponse
		err  error
	)
	resp.Valid, err = s.backend.Aprove(r.Content)
	if err != nil {
		resp.Error = &proto.Error{
			// TODO(ja): define error codes.
			Code:    0,
			Message: err.Error(),
		}
	}
	return &resp, nil
}

func (s *service) Deliver(ctx context.Context, r *proto.DeliverRequest) (*proto.DeliverResponse, error) {
	var resp proto.DeliverResponse
	if err = s.backend.Deliver(r.Content); err != nil {
		resp.Error = &proto.Error{
			// TODO(ja): define error codes.
			Code:    0,
			Message: err.Error(),
		}
	}
	return &resp, nil
}
