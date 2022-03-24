package stream

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/mock/daomock"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/kafkakit/mock/kafkamock"
	"github.com/golang/mock/gomock"
	"github.com/justin0u0/protoc-gen-grpc-sarama/pkg/saramakit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestStream(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Stream")
}

var (
	errSendMessagesUnknown = errors.New("unknown send messages error")
)

var _ = Describe("Stream", func() {
	var (
		ctx        context.Context
		controller *gomock.Controller
		videoDAO   *daomock.MockVideoDAO
		producer   *kafkamock.MockProducer
		stream     *stream
	)

	BeforeEach(func() {
		ctx = context.Background()
		controller = gomock.NewController(GinkgoT())
		videoDAO = daomock.NewMockVideoDAO(controller)
		producer = kafkamock.NewMockProducer(controller)
		stream = NewStream(videoDAO, producer)
	})

	AfterEach(func() {
		controller.Finish()
	})

	Describe("HandleVideoCreated", func() {
		var (
			id    primitive.ObjectID
			url   string
			resp  *emptypb.Empty
			err   error
			scale int32
		)

		BeforeEach(func() {
			id = primitive.NewObjectID()
			url = "https://www.test.com"
		})

		JustBeforeEach(func() {
			resp, err = stream.HandleVideoCreated(ctx, &pb.HandleVideoCreatedRequest{
				Id:    id.Hex(),
				Url:   url,
				Scale: scale,
			})
		})

		Context("scale is not presenting", func() {
			BeforeEach(func() { scale = 0 })

			When("producer send messages error", func() {
				BeforeEach(func() {
					producer.EXPECT().SendMessages(gomock.Any()).Return(errSendMessagesUnknown)
				})

				It("returns the error", func() {
					Expect(resp).To(BeNil())
					Expect(err).To(Equal(&saramakit.HandlerError{Retry: true, Err: errSendMessagesUnknown}))
				})
			})

			When("success", func() {
				BeforeEach(func() {
					producer.EXPECT().SendMessages(gomock.Any()).Times(4).Return(nil)
				})

				It("returns with no error", func() {
					Expect(resp).To(Equal(&emptypb.Empty{}))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("scale is presenting", func() {
			BeforeEach(func() { scale = 720 })

			When("video not found", func() {
				BeforeEach(func() {
					videoDAO.EXPECT().UpdateVariant(ctx, id, strconv.Itoa(int(scale)), url).Return(dao.ErrVideoNotFound)
				})

				It("returns with no error", func() {
					Expect(resp).To(BeNil())
					Expect(err).To(Equal(&saramakit.HandlerError{Retry: true, Err: dao.ErrVideoNotFound}))
				})
			})

			When("success", func() {
				BeforeEach(func() {
					videoDAO.EXPECT().UpdateVariant(ctx, id, strconv.Itoa(int(scale)), url).Return(nil)
				})

				It("returns with no error", func() {
					Expect(resp).To(Equal(&emptypb.Empty{}))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
