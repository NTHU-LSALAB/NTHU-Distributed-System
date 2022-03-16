package dao

import (
	"context"

	"github.com/go-redis/cache/v8"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ = Describe("CommentRedisDAO", func() {
	var redisCommentDAO *redisCommentDAO
	var pgCommentDAO *pgCommentDAO
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		pgCommentDAO = NewPGCommentDAO(pgClient)
		redisCommentDAO = NewRedisCommentDAO(redisClient, pgCommentDAO)
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
			comments = []*Comment{NewFakeComment(fakeVideoID), NewFakeComment(fakeVideoID), NewFakeComment(fakeVideoID)}
		})

		JustBeforeEach(func() {
			resp, err = redisCommentDAO.ListByVideoID(ctx, videoID, limit, offset)
		})

		Context("cache hit", func() {
			BeforeEach(func() {
				limit, offset = 3, 0
				videoID = comments[0].VideoID
				insertCommentsInRedis(ctx, redisCommentDAO, comments, videoID, limit, offset)
			})

			AfterEach(func() {
				deleteCommentsInRedis(ctx, redisCommentDAO, videoID, limit, offset)
			})

			When("success", func() {
				It("returns the comments with no error", func() {
					for i := range resp {
						Expect(resp[i]).To(matchComment(comments[i]))
					}
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("cache miss", func() {
			BeforeEach(func() {
				limit, offset = 2, 0
				videoID = comments[0].VideoID
				for _, comment := range comments {
					insertComment(comment)
				}
			})

			AfterEach(func() {
				for _, comment := range comments {
					deleteComment(comment.ID)
				}
				deleteCommentsInRedis(ctx, redisCommentDAO, videoID, limit, offset)
			})

			When("comments not found due to non-exist limit", func() {
				BeforeEach(func() { offset, limit = 3, 3 })

				It("returns empty list with no error", func() {
					Expect(resp).To(HaveLen(0))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("comments not found due to non-exist offset", func() {
				BeforeEach(func() { offset = 3 })

				It("returns empty list with no error", func() {
					Expect(resp).To(HaveLen(0))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("comments not found due to non-exist videoID", func() {
				BeforeEach(func() { videoID = primitive.NewObjectID().Hex() })

				It("returns empty list with no error", func() {
					Expect(resp).To(HaveLen(0))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("success", func() {
				It("returns the comments with no error", func() {
					for i := range resp {
						Expect(resp[i]).To(matchComment(comments[i]))
					}
					Expect(err).NotTo(HaveOccurred())
				})

				It("insert the comments to cache", func() {
					var getComments []*Comment
					Expect(redisCommentDAO.cache.Get(ctx, listCommentKey(videoID, limit, offset), &getComments)).NotTo(HaveOccurred())
					for i := range getComments {
						Expect(getComments[i]).To(matchComment(comments[i]))
					}
				})
			})
		})
	})
})

func insertCommentsInRedis(ctx context.Context, commentDAO *redisCommentDAO, comments []*Comment, videoID string, limit, offset int) {
	Expect(commentDAO.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   listCommentKey(videoID, limit, offset),
		Value: comments,
		TTL:   commentDAORedisCacheDuration,
	})).NotTo(HaveOccurred())
}

func deleteCommentsInRedis(ctx context.Context, commentDAO *redisCommentDAO, videoID string, limit, offset int) {
	Expect(commentDAO.cache.Delete(ctx, listCommentKey(videoID, limit, offset))).NotTo(HaveOccurred())
}
