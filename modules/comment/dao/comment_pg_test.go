package dao

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ = Describe("PGCommentDAO", func() {
	var commentDAO *pgCommentDAO
	var ctx context.Context

	BeforeEach(func() {
		commentDAO = NewPGCommentDAO(pgClient)
		ctx = context.Background()
	})

	Describe("ListByVideoID", func() {
		var (
			comments []*Comment
			videoID  string
			limit    int
			offset   int

			resp []*Comment
			err  error
		)

		BeforeEach(func() {
			fakeVideoID := primitive.NewObjectID().Hex()

			comments = []*Comment{
				NewFakeComment(fakeVideoID),
				NewFakeComment(fakeVideoID),
				NewFakeComment(fakeVideoID),
			}

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
			resp, err = commentDAO.ListByVideoID(ctx, videoID, limit, offset)
		})

		When("videos not found", func() {
			BeforeEach(func() { videoID = primitive.NewObjectID().Hex() })

			It("returns empty list with no error", func() {
				Expect(resp).To(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("success", func() {
			BeforeEach(func() { videoID = comments[0].VideoID })

			When("no limit and offset", func() {
				It("return comments with no error", func() {
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

			When("limit = 1 and offset = 1", func() {
				BeforeEach(func() { limit, offset = 1, 1 })

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

			resp uuid.UUID
			err  error
		)

		BeforeEach(func() {
			comment = NewFakeComment("")
			comment.ID = uuid.Nil
		})

		AfterEach(func() {
			deleteComment(comment.ID)
		})

		JustBeforeEach(func() {
			resp, err = commentDAO.Create(ctx, comment)
		})

		When("success", func() {
			It("returns the new comment ID with no error", func() {
				Expect(resp).NotTo(Equal(uuid.Nil))
				Expect(err).NotTo(HaveOccurred())
			})

			It("inserts the comment", func() {
				var getComment Comment

				_, err := pgClient.QueryOne(&getComment, "SELECT * FROM comments WHERE id = ?", comment.ID)

				Expect(&getComment).To(matchComment(comment))
				Expect(err).NotTo(HaveOccurred())
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
			comment = NewFakeComment("")
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

				_, err := pgClient.QueryOne(&getComment, "SELECT * FROM comments WHERE id = ?", comment.ID)

				Expect(err).NotTo(HaveOccurred())
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
			comment = NewFakeComment("")

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

			It("deletes the comment", func() {
				var getComment, emptyComment Comment

				_, err := pgClient.QueryOne(&getComment, "SELECT * FROM comments WHERE id = ?", id)

				Expect(getComment).To(Equal(emptyComment))
				Expect(err).To(MatchError(pg.ErrNoRows))
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
			fakeVideoID := primitive.NewObjectID().Hex()

			comments = []*Comment{
				NewFakeComment(fakeVideoID),
				NewFakeComment(fakeVideoID),
				NewFakeComment(fakeVideoID),
			}

			for _, comment := range comments {
				insertComment(comment)
			}
		})

		JustBeforeEach(func() {
			err = commentDAO.DeleteByVideoID(ctx, videoID)
		})

		When("success", func() {
			BeforeEach(func() { videoID = comments[0].VideoID })

			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes those comments", func() {
				var getComments []*Comment

				_, err := pgClient.QueryOne(&getComments, "SELECT * FROM comments WHERE video_id = ?", videoID)

				Expect(getComments).To(HaveLen(0))
				Expect(err).To(MatchError(pg.ErrNoRows))
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

func matchComment(comment *Comment) types.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"ID":      Equal(comment.ID),
		"VideoID": Equal(comment.VideoID),
		"Content": Equal(comment.Content),
	}))
}
