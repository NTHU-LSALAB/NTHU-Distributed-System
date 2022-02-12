package service

import (
	"context"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
)

type service struct {
	pb.UnimplementedVideoServer
}

func (s *service) Healthz(ctx context.Context, req *pb.HealthzRequest) (*pb.HealthzResponse, error) {
	return &pb.HealthzResponse{Status: "ok"}, nil
}
