package dao

import (
	"context"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/postgreskit"
)

type commentPGDAO struct {
	pgClient *postgreskit.PGClient
}

var _ CommentDAO = (*commentPGDAO)(nil)

func NewCommentPGDAO(pgClient *postgreskit.PGClient) *commentPGDAO {
	return &commentPGDAO{
		pgClient: pgClient,
	}
}

func (dao *commentPGDAO) List(ctx context.Context, video_id string, limit, skip int) ([]*Comment, error) {
	var comments []*Comment
	query := dao.pgClient.ModelContext(ctx, &comments).
		Where("video_id=?", video_id).
		Limit(limit).
		Offset(skip).
		Order("updated_at DESC")
	if err := query.Select(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (dao *commentPGDAO) Create(ctx context.Context, comment *Comment) error {
	if _, err := dao.pgClient.ModelContext(ctx, comment).Insert(); err != nil {
		return err
	}

	return nil
}

func (dao *commentPGDAO) Update(ctx context.Context, comment *Comment) error {
	if res, err := dao.pgClient.ModelContext(ctx, comment).WherePK().Update(); err != nil {
		return err
	} else if res.RowsAffected() == 0 {
		return ErrCommentNotFound
	}

	return nil
}

func (dao *commentPGDAO) Delete(ctx context.Context, id int32) error {
	var comment *Comment
	if _, err := dao.pgClient.ModelContext(ctx, comment).Where("_id = ?", id).Delete(); err != nil {
		return err
	}

	return nil
}

// delete all comments when the video deleted
func (dao *commentPGDAO) DeleteComments(ctx context.Context, video_id string) error {
	var comment *Comment
	if _, err := dao.pgClient.ModelContext(ctx, comment).Where("video_id=?", video_id).Delete(); err != nil {
		return err
	}

	return nil
}
