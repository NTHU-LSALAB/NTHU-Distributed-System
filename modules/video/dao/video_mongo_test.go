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

			Expect(videoDAO.collection.InsertOne(ctx, video)).To(Equal(&mongo.InsertOneResult{
				InsertedID: video.ID,
			}))
		})

		AfterEach(func() {
			Expect(videoDAO.collection.DeleteOne(ctx, bson.M{"_id": video.ID})).To(Equal(&mongo.DeleteResult{
				DeletedCount: 1,
			}))
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
})
