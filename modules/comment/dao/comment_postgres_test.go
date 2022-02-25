package dao

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
			fake_video_id := primitive.NewObjectID().Hex()

			for _, comment := range comments {
				comment.VideoID = fake_video_id
				insertComment(comment)
			}
		})

		AfterEach(func() {
			deletCommentByVideoID(comments[0].VideoID)
		})

		JustBeforeEach(func() {
			resp, err = commentDAO.List(ctx, video_id, limit, skip)
		})

		When("video not found", func() {
			BeforeEach(func() { video_id = primitive.NewObjectID().Hex() })

			It("returns comment not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrVideoNotFound))
			})
		})

		Context("success", func() {
			BeforeEach(func() { video_id = comments[0].VideoID })

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

			res uuid.UUID
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
			res, err = commentDAO.Create(ctx, comment)
		})

		When("success", func() {
			It("returns the new comment ID with no error", func() {
				Expect(res).NotTo(Equal(uuid.Nil))
				Expect(err).NotTo(HaveOccurred())
			})

			It("inserts the comment", func() {
				var getComment Comment
				query := "SELECT * FROM comments WHERE id = ?"
				_, err := pgClient.DB.QueryOne(&getComment, query, comment.ID)

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
				var getComment Comment
				query := "SELECT * FROM comments WHERE id = ?"
				_, err := pgClient.DB.QueryOne(&getComment, query, comment.ID)

				Expect(err).NotTo(HaveOccurred())
				Expect(getComment.Content).To(Equal(content))
				Expect(getComment).To(Equal(comment))
			})
		})
	})

	Describe("Delete", func() {
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

		JustBeforeEach(func() {
			err = commentDAO.Delete(ctx, id)
		})

		When("comment not found", func() {
			BeforeEach(func() { comment.ID = uuid.New() })

			AfterEach(func() {
				deleteComment(id)
			})

			It("returns comment not found error", func() {
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		When("success", func() {
			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeleteByVideoID", func() {
		var (
			comments []*Comment
			video_id string

			err error
		)

		BeforeEach(func() {
			comments = []*Comment{NewFakeComment(), NewFakeComment(), NewFakeComment()}
			fake_video_id := primitive.NewObjectID().Hex()

			for _, comment := range comments {
				comment.VideoID = fake_video_id
				insertComment(comment)
			}
		})

		JustBeforeEach(func() {
			err = commentDAO.DeleteByVideoID(ctx, video_id)
		})

		When("video not found", func() {
			BeforeEach(func() { video_id = primitive.NewObjectID().Hex() })

			AfterEach(func() {
				deletCommentByVideoID(comments[0].VideoID)
			})

			It("returns comment not found error", func() {
				Expect(err).To(MatchError(ErrVideoNotFound))
			})
		})

		When("success", func() {
			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
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

func deletCommentByVideoID(video_id string) {
	query := "DELETE FROM comments WHERE video_id = ?"

	pgExec(query, video_id)
}
