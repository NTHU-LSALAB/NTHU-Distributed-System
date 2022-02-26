package dao

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ = Describe("CommentPostgresDAO", func() {
	var commentDAO *pgCommentDAO
	var ctx context.Context

	BeforeEach(func() {
		commentDAO = NewPGCommentDAO(pgClient)
		ctx = context.Background()
	})

	Describe("List", func() {
		var (
			comments []*Comment
			videoID  string
			limit    int
			skip     int

			resp []*Comment
			err  error
		)

		BeforeEach(func() {
			comments = []*Comment{NewFakeComment(), NewFakeComment(), NewFakeComment()}
			fakeVideoID := primitive.NewObjectID().Hex()

			for i := 0; i < len(comments); i++ {
				comments[i].VideoID = fakeVideoID

				insertComment(comments[i])
			}
		})

		AfterEach(func() {
			deletCommentByVideoID(comments[0].VideoID)
		})

		JustBeforeEach(func() {
			resp, err = commentDAO.List(ctx, videoID, limit, skip)
		})

		When("video not found", func() {
			BeforeEach(func() { videoID = primitive.NewObjectID().Hex() })

			It("returns video not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("success", func() {
			BeforeEach(func() { videoID = comments[0].VideoID })

			When("no limit and offset", func() {
				It("return comments with no offer", func() {
					for i := range resp {
						Expect(resp[i]).To(matchComment(comments[i]))
					}
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("limit = 1", func() {
				BeforeEach(func() { limit = 1 })

				It("returns the first comment with no error", func() {
					Expect(resp[0]).To(matchComment(comments[0]))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("limit = 1 and skip = 1", func() {
				BeforeEach(func() { limit, skip = 1, 1 })

				It("returns the second comment with no error", func() {
					Expect(resp[0]).To(matchComment(comments[1]))
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
				Expect(&getComment).To(Equal(comment))
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
				Expect(&getComment).To(matchComment(comment))
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

			insertComment(comment)
		})

		JustBeforeEach(func() {
			err = commentDAO.Delete(ctx, id)
		})

		When("comment not found", func() {
			BeforeEach(func() { id = uuid.New() })

			AfterEach(func() {
				deleteComment(comment.ID)
			})

			It("returns comment not found error", func() {
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		When("success", func() {
			BeforeEach(func() { id = comment.ID })

			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeleteByVideoID", func() {
		var (
			comments []*Comment
			videoID  string

			err error
		)

		BeforeEach(func() {
			comments = []*Comment{NewFakeComment(), NewFakeComment(), NewFakeComment()}
			fakeVideoID := primitive.NewObjectID().Hex()

			for i := 0; i < len(comments); i++ {
				comments[i].VideoID = fakeVideoID
				insertComment(comments[i])
			}
		})

		JustBeforeEach(func() {
			err = commentDAO.DeleteByVideoID(ctx, videoID)
		})

		When("video not found", func() {
			BeforeEach(func() { videoID = primitive.NewObjectID().Hex() })

			AfterEach(func() {
				deletCommentByVideoID(comments[0].VideoID)
			})

			It("returns comment not found error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("success", func() {
			BeforeEach(func() { videoID = comments[0].VideoID })
			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

func insertComment(comment *Comment) {
	query := "INSERT INTO comments (id, video_id, content) VALUES (?, ?, ?);"

	pgExec(query, comment.ID, comment.VideoID, comment.Content)
}

func deleteComment(id uuid.UUID) {
	query := "DELETE FROM comments WHERE id = ?;"

	pgExec(query, id)
}

func deletCommentByVideoID(videoID string) {
	query := "DELETE FROM comments WHERE video_id = ?;"

	pgExec(query, videoID)
}

func matchComment(comment *Comment) types.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"ID":      Equal(comment.ID),
		"VideoID": Equal(comment.VideoID),
		"Content": Equal(comment.Content),
	}))
}
