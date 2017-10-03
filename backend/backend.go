package backend

import (
	"log"
	"net"

	"github.com/WiseGrowth/pigeon/pigeon"
	"github.com/WiseGrowth/pigeon/pigeon/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func ListenAndServe(addr pigeon.NetAddr, backend pigeon.Backend) error {
	lis, err := net.Listen("tcp", string(addr))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	proto.RegisterBackendServiceServer(s, &service{backend})

	return s.Serve(lis)
}

type service struct {
	backend pigeon.Backend
}

func (s *service) Approve(ctx context.Context, r *proto.ApproveRequest) (*proto.ApproveResponse, error) {
	var (
		resp proto.ApproveResponse
		err  error
	)
	resp.Valid, err = s.backend.Approve(r.Content)
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
	if err := s.backend.Deliver(r.Content); err != nil {
		resp.Error = &proto.Error{
			// TODO(ja): define error codes.
			Code:    0,
			Message: err.Error(),
		}
	}
	return &resp, nil
}
