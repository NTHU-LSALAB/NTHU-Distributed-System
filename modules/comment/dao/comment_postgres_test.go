package dao

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommentPostgresDAO", func() {
	var commentDAO *commentPGDAO
	var ctx context.Context

	BeforeEach(func() {
		commentDAO = NewCommentPGDAO(pgClient)
		ctx = context.Background()
	})

	Describe("Get", func() {
		var (
			comment *Comment
			id      uuid.UUID

			resp *Comment
			err  error
		)

		BeforeEach(func() {
			comment = NewFakeComment()

			insertComment(ctx, commentDAO, comment)
		})

		AfterEach(func() {

		})
	})
})

func insertComment(ctx context.Context, commentDAO *commentPGDAO, comment *Comment) {
}
