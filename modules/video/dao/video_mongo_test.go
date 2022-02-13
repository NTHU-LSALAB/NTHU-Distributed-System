package dao

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ = Describe("VideoMongoDAO", func() {
	var videoDAO *videoMongoDAO

	BeforeEach(func() {
		videoDAO = NewVideoMongoDAO(mongoClient.Database().Collection("video"))
	})

	Describe("GetVideo", func() {
		var (
			ctx   context.Context
			video *Video
			id    primitive.ObjectID

			resp *Video
			err  error
		)

		BeforeEach(func() {
			ctx = context.Background()
			video = newFakeVideo()

			Expect(videoDAO.collection.InsertOne(ctx, video)).
				To(Equal(&mongo.InsertOneResult{InsertedID: video.ID}))
		})

		AfterEach(func() {
			Expect(videoDAO.collection.DeleteOne(ctx, bson.M{"_id": video.ID})).
				To(Equal(&mongo.DeleteResult{DeletedCount: 1}))
		})

		JustBeforeEach(func() {
			resp, err = videoDAO.Get(ctx, id)
		})

		When("video not found", func() {
			BeforeEach(func() { id = primitive.NewObjectID() })

			It("returns video not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrVideoNotFound))
			})
		})

		When("success", func() {
			BeforeEach(func() { id = video.ID })

			It("returns the video with no error", func() {
				Expect(resp).To(Equal(video))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("ListVideo", func() {
		var (
			ctx    context.Context
			videos []*Video
			limit  int64
			skip   int64

			resp []*Video
			err  error
		)

		BeforeEach(func() {
			ctx = context.Background()

			videos = []*Video{newFakeVideo(), newFakeVideo(), newFakeVideo()}

			for _, video := range videos {
				Expect(videoDAO.collection.InsertOne(ctx, video)).
					To(Equal(&mongo.InsertOneResult{InsertedID: video.ID}))
			}
		})

		AfterEach(func() {
			for _, video := range videos {
				Expect(videoDAO.collection.DeleteOne(ctx, bson.M{"_id": video.ID})).
					To(Equal(&mongo.DeleteResult{DeletedCount: 1}))
			}
		})

		JustBeforeEach(func() {
			resp, err = videoDAO.List(ctx, limit, skip)
		})

		Context("success", func() {
			When("no limit and offset", func() {
				It("returns videos with no error", func() {
					Expect(resp).To(Equal(videos))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("limit = 1", func() {
				BeforeEach(func() { limit = 1 })

				It("returns the first video with no error", func() {
					Expect(resp).To(Equal(videos[:1]))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("limit = 1 and skip = 1", func() {
				BeforeEach(func() { limit, skip = 1, 1 })

				It("returns the second video with no error", func() {
					Expect(resp).To(Equal(videos[1:2]))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("CreateVideo", func() {
		var (
			ctx   context.Context
			video *Video

			err error

			getVideo Video
		)

		BeforeEach(func() {
			ctx = context.Background()

			video = newFakeVideo()
			video.ID = primitive.NilObjectID
		})

		AfterEach(func() {
			Expect(videoDAO.collection.DeleteOne(ctx, bson.M{"_id": video.ID})).
				To(Equal(&mongo.DeleteResult{DeletedCount: 1}))
		})

		JustBeforeEach(func() {
			err = videoDAO.Create(ctx, video)

			Expect(
				videoDAO.collection.FindOne(ctx, bson.M{"_id": video.ID}).Decode(&getVideo),
			).NotTo(HaveOccurred())
		})

		When("success", func() {
			It("returns the new video ID with no error", func() {
				Expect(video.ID).NotTo(Equal(primitive.NilObjectID))
				Expect(err).NotTo(HaveOccurred())
			})

			It("inserts the document successfully", func() {
				Expect(&getVideo).To(Equal(video))
			})
		})
	})

})
