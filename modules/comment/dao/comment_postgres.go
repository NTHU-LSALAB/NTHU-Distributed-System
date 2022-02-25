package dao

import (
	"context"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/pgkit"
	"github.com/google/uuid"
)

type commentPGDAO struct {
	client *pgkit.PGClient
}

var _ CommentDAO = (*commentPGDAO)(nil)

func NewCommentPGDAO(pgClient *pgkit.PGClient) *commentPGDAO {
	return &commentPGDAO{
		client: pgClient,
	}
}

func (dao *commentPGDAO) List(ctx context.Context, videoID string, limit, offset int) ([]*Comment, error) {
	var comments []*Comment
	query := dao.client.ModelContext(ctx, &comments).
		Where("video_id = ?", videoID).
		Limit(limit).
		Offset(offset).
		Order("updated_at DESC")
	if err := query.Select(); err != nil {
		return nil, ErrVideoNotFound
	}

	return comments, nil
}

func (dao *commentPGDAO) Create(ctx context.Context, comment *Comment) (uuid.UUID, error) {
	if _, err := dao.client.ModelContext(ctx, comment).Insert(); err != nil {
		return uuid.Nil, err
	}

	return comment.ID, nil
}

func (dao *commentPGDAO) Update(ctx context.Context, comment *Comment) error {
	if res, err := dao.client.ModelContext(ctx, comment).Column("content").WherePK().Update(); err != nil {
		return err
	} else if res.RowsAffected() == 0 {
		return ErrCommentNotFound
	}

	return nil
}

func (dao *commentPGDAO) Delete(ctx context.Context, id uuid.UUID) error {
	if res, err := dao.client.ModelContext(ctx, &Comment{ID: id}).WherePK().Delete(); err != nil {
		return err
	} else if res.RowsAffected() == 0 {
		return ErrCommentNotFound
	}

	return nil
}

// delete all comments when the video deleted
func (dao *commentPGDAO) DeleteByVideoID(ctx context.Context, videoID string) error {
	var comment *Comment
	if _, err := dao.client.ModelContext(ctx, comment).Where("video_id = ?", videoID).Delete(); err != nil {
		return ErrVideoNotFound
	}

	return nil
}
