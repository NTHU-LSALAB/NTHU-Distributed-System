package otelkit

import (
	"context"
	"net/http"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"github.com/prometheus/common/expfmt"
	"google.golang.org/grpc"
)

var _ = Describe("PrometheusServiceMeter", func() {
	var (
		ctx         context.Context
		conf        *PrometheusServiceMeterConfig
		meter       *PrometheusServiceMeter
		interceptor grpc.UnaryServerInterceptor
	)

	BeforeEach(func() {
		ctx = context.Background()
		ctx = logkit.NewNopLogger().WithContext(ctx)

		conf = &PrometheusServiceMeterConfig{
			Addr:                ":52222",
			Path:                "/metrics",
			Name:                "test_prometheus_service_meter",
			HistogramBoundaries: []float64{10, 100},
		}

		meter = NewPrometheusServiceMeter(ctx, conf)
		time.Sleep(50 * time.Millisecond) // wait prometheus exporter server to start

		interceptor = meter.UnaryServerInterceptor()
	})

	AfterEach(func() {
		Expect(meter.Close()).NotTo(HaveOccurred())
	})

	Context("single handler success", func() {
		var (
			handler      grpc.UnaryHandler
			responseTime time.Duration
			handlerReq   interface{}
			handlerResp  interface{}
			resp         interface{}
			err          error
		)

		BeforeEach(func() {
			handlerReq = "fake request"
			handlerResp = "fake response"
		})

		JustBeforeEach(func() {
			handler = func(ctx context.Context, req interface{}) (interface{}, error) {
				time.Sleep(responseTime)
				return handlerResp, nil
			}

			resp, err = interceptor(ctx, handlerReq, &grpc.UnaryServerInfo{
				FullMethod: "test_handler_success",
			}, handler)
		})

		When("handler takes 5ms to finish", func() {
			BeforeEach(func() { responseTime = 5 * time.Millisecond })

			It("success", func() {
				validateMetric(ctx, conf, 1, responseTime)
			})

			It("does not change handler response", func() {
				Expect(resp).To(Equal(handlerResp))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("handler takes 50ms to finish", func() {
			BeforeEach(func() { responseTime = 50 * time.Millisecond })

			It("success", func() {
				validateMetric(ctx, conf, 1, responseTime)
			})

			It("does not change handler response", func() {
				Expect(resp).To(Equal(handlerResp))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("handler takes 150ms to finish", func() {
			BeforeEach(func() { responseTime = 150 * time.Millisecond })

			It("success", func() {
				validateMetric(ctx, conf, 1, responseTime)
			})

			It("does not change handler response", func() {
				Expect(resp).To(Equal(handlerResp))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

func validateMetric(ctx context.Context, conf *PrometheusServiceMeterConfig, handlerCallCount int, responseTime time.Duration) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+conf.Addr+conf.Path, http.NoBody)
	Expect(err).NotTo(HaveOccurred())

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())

	defer func() {
		Expect(resp.Body.Close()).NotTo(HaveOccurred())
	}()

	var parser expfmt.TextParser
	mfs, err := parser.TextToMetricFamilies(resp.Body)
	Expect(err).NotTo(HaveOccurred())

	requestMF := mfs["request"]
	Expect(requestMF.GetName()).To(Equal("request"))
	Expect(requestMF.GetType().String()).To(Equal("COUNTER"))
	Expect(requestMF.GetMetric()).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
		"Counter": PointTo(MatchFields(IgnoreExtras, Fields{
			"Value": PointTo(Equal(float64(handlerCallCount))),
		})),
	}))))

	responseTimeMF := mfs["response_time"]
	Expect(responseTimeMF.GetName()).To(Equal("response_time"))
	Expect(responseTimeMF.GetType().String()).To(Equal("HISTOGRAM"))
	Expect(responseTimeMF.GetMetric()).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
		"Histogram": PointTo(MatchFields(IgnoreExtras, Fields{
			"SampleCount": PointTo(Equal(uint64(handlerCallCount))),
			"Bucket":      matchBucket(conf, handlerCallCount, responseTime),
		})),
	}))))
}

func matchBucket(conf *PrometheusServiceMeterConfig, handlerCallCount int, responseTime time.Duration) types.GomegaMatcher {
	matcher := make([]types.GomegaMatcher, 0, len(conf.HistogramBoundaries))

	for _, bound := range conf.HistogramBoundaries {
		count := 0
		if float64(responseTime.Milliseconds()) <= bound {
			count = handlerCallCount
		}

		matcher = append(matcher, PointTo(MatchFields(IgnoreExtras, Fields{
			"CumulativeCount": PointTo(Equal(uint64(count))),
			"UpperBound":      PointTo(Equal(bound)),
		})))
	}

	return ContainElements(matcher)
}
