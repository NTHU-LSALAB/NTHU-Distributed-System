package dao

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Comment struct {
	ID        int32     `pg:"_id,omitempty"`
	VideoID   string    `pg:"video_id,omitempty"`
	Content   string    `pg:"content,omitempty"`
	CreatedAt time.Time `pg:"created_at,omitempty"`
	UpdatedAt time.Time `pg:"updated_at,omitempty"`
}

func (c *Comment) ToProto() *pb.CommentInfo {
	return &pb.CommentInfo{
		Id:        c.ID,
		VideoId:   c.VideoID,
		Content:   c.Content,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

type CommentDAO interface {
	List(ctx context.Context, video_id string, limit, skip int) ([]*Comment, error)
	Create(ctx context.Context, comment *Comment) error
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id int32) error
	DeleteComments(ctx context.Context, video_id string) error
}

var (
	ErrCommentNotFound = errors.New("Comment not found")
)

func NewFakeComment() *Comment {
	var id int32 = int32(rand.Int())
	video_id := primitive.NewObjectID()

	return &Comment{
		ID:      id,
		VideoID: video_id.Hex(),
		Content: "comment test",
	}
}
