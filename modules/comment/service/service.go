package service

import (
	"context"
	"errors"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	videopb "github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/google/uuid"
)

type service struct {
	pb.UnimplementedCommentServer

	commentDAO  dao.CommentDAO
	videoClient videopb.VideoClient
}

func NewService(commentDAO dao.CommentDAO, videoClient videopb.VideoClient) *service {
	return &service{
		commentDAO:  commentDAO,
		videoClient: videoClient,
	}
}

func (s *service) Healthz(ctx context.Context, req *pb.HealthzRequest) (*pb.HealthzResponse, error) {
	return &pb.HealthzResponse{Status: "ok"}, nil
}

func (s *service) ListComment(ctx context.Context, req *pb.ListCommentRequest) (*pb.ListCommentResponse, error) {
	comments, err := s.commentDAO.ListByVideoID(ctx, req.GetVideoId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}

	pbComments := make([]*pb.CommentInfo, 0, len(comments))
	for _, comment := range comments {
		pbComments = append(pbComments, comment.ToProto())
	}

	return &pb.ListCommentResponse{Comments: pbComments}, nil
}

func (s *service) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CreateCommentResponse, error) {
	if _, err := s.videoClient.GetVideo(ctx, &videopb.GetVideoRequest{
		Id: req.GetVideoId(),
	}); err != nil {
		return nil, err
	}

	comment := &dao.Comment{
		VideoID: req.GetVideoId(),
		Content: req.GetContent(),
	}

	commentID, err := s.commentDAO.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	return &pb.CreateCommentResponse{Id: commentID.String()}, nil
}

func (s *service) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.UpdateCommentResponse, error) {
	commentID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, ErrInvalidUUID
	}

	comment := &dao.Comment{
		ID:      commentID,
		Content: req.GetContent(),
	}
	if err := s.commentDAO.Update(ctx, comment); err != nil {
		if errors.Is(err, dao.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}

		return nil, err
	}

	return &pb.UpdateCommentResponse{
		Comment: comment.ToProto(),
	}, nil
}

func (s *service) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	commentID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, ErrInvalidUUID
	}

	if err := s.commentDAO.Delete(ctx, commentID); err != nil {
		if errors.Is(err, dao.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}

		return nil, err
	}

	return &pb.DeleteCommentResponse{}, nil
}

func (s *service) DeleteCommentByVideoID(ctx context.Context, req *pb.DeleteCommentByVideoIDRequest) (*pb.DeleteCommentByVideoIDResponse, error) {
	if err := s.commentDAO.DeleteByVideoID(ctx, req.GetVideoId()); err != nil {
		return nil, err
	}

	return &pb.DeleteCommentByVideoIDResponse{}, nil
}
