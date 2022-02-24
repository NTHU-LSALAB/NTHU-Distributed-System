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

	Describe("List", func() {
		var (
			comments []*Comment
			video_id string
			limit    int
			skip     int

			resp []*Comment
			err  error
		)

		BeforeEach(func() {
			comments = []*Comment{NewFakeComment(), NewFakeComment(), NewFakeComment()}

			for _, comment := range comments {
				insertComment(comment)
			}
		})

		AfterEach(func() {
			for _, comment := range comments {
				deleteComment(comment.ID)
			}
		})

		JustBeforeEach(func() {
			resp, err = commentDAO.List(ctx, video_id, limit, skip)
		})

		When("comment not found", func() {
			BeforeEach(func() { video_id = uuid.NewString() })

			It("returns comment not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		Context("success", func() {
			When("no limit and offset", func() {
				It("return comments with no offer", func() {
					Expect(resp).To(Equal(comments))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("limit = 1", func() {
				BeforeEach(func() { limit = 1 })

				It("returns the first comment with no error", func() {
					Expect(resp).To(Equal(comments[:1]))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("limit = 1 and skip = 1", func() {
				BeforeEach(func() { limit, skip = 1, 1 })

				It("returns the second comment with no error", func() {
					Expect(resp).To(Equal(comments[1:2]))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("Create", func() {
		var (
			comment *Comment

			err error
		)

		BeforeEach(func() {
			comment = NewFakeComment()
			comment.ID = uuid.Nil
		})

		AfterEach(func() {
			deleteComment(comment.ID)
		})

		JustBeforeEach(func() {
			err = commentDAO.Create(ctx, comment)
		})

		When("success", func() {
			It("returns the new comment ID with no error", func() {
				Expect(comment.ID).NotTo(Equal(uuid.Nil))
				Expect(err).NotTo(HaveOccurred())
			})

			It("inserts the comment", func() {
				query := "SELECT * FROM comments WHERE id = ?"
				getComment, err := pgClient.Exec(query, comment.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(getComment).To(Equal(comment))
			})
		})
	})

	Describe("Update", func() {
		var (
			comment *Comment
			id      uuid.UUID

			err error
		)

		BeforeEach(func() {
			comment = NewFakeComment()
			id = comment.ID

			insertComment(comment)
		})

		AfterEach(func() {
			deleteComment(id)
		})

		JustBeforeEach(func() {
			err = commentDAO.Update(ctx, comment)
		})

		When("comment not found", func() {
			BeforeEach(func() { comment.ID = uuid.New() })

			It("returns comment not found error", func() {
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		When("success", func() {
			var content string

			BeforeEach(func() {
				content = "comment update test"
				comment.Content = content
			})

			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the comment", func() {
				query := "SELECT * FROM comments WHERE id = ?"
				res, err := pgClient.Exec(query, comment.ID)
				Expect(res.Model).To(Equal(content))
				Expect(getComment).To(Equal(comment))
			})
		})
	})
})

func insertComment(comment *Comment) {
	query := "INSERT INTO comments (id, video_id, content) VALUES (?, ?, ?)"

	pgExec(query, comment.ID, comment.VideoID, comment.Content)
}

func deleteComment(id uuid.UUID) {
	query := "DELETE FROM comments WHERE id = ?"

	pgExec(query, id)
}
