package stream

import (
	"errors"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/kafkakit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type handler struct {
	*VideoCreatedHandler
}

func NewHandler(server pb.VideoStreamServer, logger *logkit.Logger) *handler {
	return &handler{
		VideoCreatedHandler: &VideoCreatedHandler{
			server:      server,
			unmarshaler: &proto.UnmarshalOptions{},
			logger:      logger.With(zap.String("Handler", "VideoCreatedHandler")),
		},
	}
}

type VideoCreatedHandler struct {
	server      pb.VideoStreamServer
	unmarshaler *proto.UnmarshalOptions
	logger      *logkit.Logger
}

var _ sarama.ConsumerGroupHandler = (*VideoCreatedHandler)(nil)

func (h *VideoCreatedHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *VideoCreatedHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *VideoCreatedHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var req pb.HandleVideoCreatedRequest

		if err := h.unmarshaler.Unmarshal(msg.Value, &req); err != nil {
			// unretryable failed, log error then skip and consume the message
			h.logger.Error("failed to unmarshal message", zap.Error(err))

			continue
		}

		if _, err := h.server.HandleVideoCreated(sess.Context(), &req); err != nil {
			var e kafkakit.HandlerError

			if ok := errors.As(err, &e); ok && e.Retry {
				h.logger.Error("failed to HandleVideoCreated and the error is retryable", zap.Error(err))

				return nil
			}

			h.logger.Error("failed to HandleVideoCreated and the error is unretryable", zap.Error(err))
		}

		// mark message as completed
		sess.MarkMessage(msg, "")
	}

	return nil
}
