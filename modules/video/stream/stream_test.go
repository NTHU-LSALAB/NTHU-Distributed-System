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
	. "github.com/onsi/ginkgo"
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
		controller *gomock.Controller
		videoDAO   *daomock.MockVideoDAO
		producer   *kafkamock.MockProducer
		stream     *stream
		ctx        context.Context
	)

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		videoDAO = daomock.NewMockVideoDAO(controller)
		producer = kafkamock.NewMockProducer(controller)
		stream = NewStream(videoDAO, producer)
		ctx = context.Background()
	})

	AfterEach(func() {
		controller.Finish()
	})

	Describe("HandleVideoCreated", func() {
		var (
			req   *pb.HandleVideoCreatedRequest
			resp  *emptypb.Empty
			id    primitive.ObjectID
			err   error
			url   string
			scale int32
		)

		BeforeEach(func() {
			id = primitive.NewObjectID()
			url = "https://www.test.com"
		})

		JustBeforeEach(func() {
			resp, err = stream.HandleVideoCreated(ctx, req)
		})

		Context("Scale = 0", func() {
			BeforeEach(func() {
				scale = 0
				req = &pb.HandleVideoCreatedRequest{
					Id:    id.Hex(),
					Url:   url,
					Scale: scale,
				}
			})

			When("Producer send messages error", func() {
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

		Context("Scale != 0", func() {
			BeforeEach(func() {
				scale = 720
				req = &pb.HandleVideoCreatedRequest{
					Id:    id.Hex(),
					Url:   url,
					Scale: scale,
				}
			})

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
