package service

import (
	"context"
	"errors"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service struct {
	pb.UnimplementedVideoServer

	videoDAO dao.VideoDAO
	logger   *logkit.Logger
}

func (s *service) Healthz(ctx context.Context, req *pb.HealthzRequest) (*pb.HealthzResponse, error) {
	return &pb.HealthzResponse{Status: "ok"}, nil
}

func (s *service) GetVideo(ctx context.Context, req *pb.GetVideoRequest) (*pb.GetVideoResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, ErrInvalidObjectID
	}

	video, err := s.videoDAO.Get(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrVideoNotFound) {
			return nil, ErrVideoNotFound
		}

		return nil, err
	}

	return &pb.GetVideoResponse{Video: video.ToProto()}, nil
}

func (s *service) ListVideo(ctx context.Context, req *pb.ListVideoRequest) (*pb.ListVideoResponse, error) {
	videos, err := s.videoDAO.List(ctx, req.GetLimit(), req.GetSkip())
	if err != nil {
		return nil, err
	}

	pbVideos := make([]*pb.VideoInfo, len(videos))
	for _, video := range videos {
		pbVideos = append(pbVideos, video.ToProto())
	}

	return &pb.ListVideoResponse{Videos: pbVideos}, nil
}

func (s *service) UploadVideo(stream pb.Video_UploadVideoServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadVideo not implemented")
}

func (s *service) DeleteVideo(ctx context.Context, req *pb.DeleteVideoRequest) (*pb.DeleteVideoResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, ErrInvalidObjectID
	}

	if err := s.videoDAO.Delete(ctx, id); err != nil {
		if errors.Is(err, dao.ErrVideoNotFound) {
			return nil, ErrVideoNotFound
		}

		return nil, err
	}

	return &pb.DeleteVideoResponse{}, nil
}
