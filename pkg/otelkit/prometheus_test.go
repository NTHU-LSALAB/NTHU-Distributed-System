package otelkit

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	prompb "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"google.golang.org/grpc"
)

var errUnknown = errors.New("unknown error")

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
			handlerResp  interface{}
			resp         interface{}
			err          error
		)

		BeforeEach(func() {
			handlerResp = "fake response"
		})

		JustBeforeEach(func() {
			handler = func(ctx context.Context, req interface{}) (interface{}, error) {
				time.Sleep(responseTime)
				return handlerResp, nil
			}

			resp, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{
				FullMethod: "test_handler_success",
			}, handler)
		})

		for _, t := range []time.Duration{5 * time.Millisecond, 50 * time.Millisecond, 150 * time.Millisecond} {
			t := t

			When("handler takes "+responseTime.String()+" to finish", func() {
				BeforeEach(func() { responseTime = t })

				It("record metrics correctly", func() {
					validateCounter(ctx, conf, "request", 1)
					validateHistogram(ctx, conf, "response_time", []float64{float64(responseTime.Milliseconds())})
				})

				It("does not change handler response", func() {
					Expect(resp).To(Equal(handlerResp))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		}
	})

	Context("single handler fail", func() {
		var (
			handler      grpc.UnaryHandler
			responseTime time.Duration
			handlerResp  interface{}
			resp         interface{}
			err          error
		)

		BeforeEach(func() {
			handlerResp = "fake response"
		})

		JustBeforeEach(func() {
			handler = func(ctx context.Context, req interface{}) (interface{}, error) {
				time.Sleep(responseTime)
				return handlerResp, errUnknown
			}

			resp, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{
				FullMethod: "test_handler_fail_and_success",
			}, handler)
		})

		for _, t := range []time.Duration{5 * time.Millisecond, 50 * time.Millisecond, 150 * time.Millisecond} {
			t := t

			When("handler takes "+responseTime.String()+" to finish", func() {
				BeforeEach(func() { responseTime = t })

				It("record metrics correctly", func() {
					validateCounter(ctx, conf, "request", 1)
					validateCounter(ctx, conf, "error_request", 1)
					validateHistogram(ctx, conf, "response_time", []float64{float64(responseTime.Milliseconds())})
				})

				It("does not change handler response", func() {
					Expect(resp).To(Equal(handlerResp))
					Expect(err).To(MatchError(errUnknown))
				})
			})
		}
	})
})

func validateCounter(ctx context.Context, conf *PrometheusServiceMeterConfig, name string, count int) {
	mf := parseMetric(ctx, conf, name)

	Expect(mf.GetName()).To(Equal(name))
	Expect(mf.GetType().String()).To(Equal("COUNTER"))
	Expect(mf.GetMetric()).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
		"Counter": PointTo(MatchFields(IgnoreExtras, Fields{
			"Value": PointTo(Equal(float64(count))),
		})),
	}))))
}

func validateHistogram(ctx context.Context, conf *PrometheusServiceMeterConfig, name string, values []float64) {
	mf := parseMetric(ctx, conf, name)

	Expect(mf.GetName()).To(Equal(name))
	Expect(mf.GetType().String()).To(Equal("HISTOGRAM"))
	Expect(mf.GetMetric()).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
		"Histogram": PointTo(MatchFields(IgnoreExtras, Fields{
			"SampleCount": PointTo(Equal(uint64(len(values)))),
			"Bucket":      matchBucket(conf, values),
		})),
	}))))
}

func parseMetric(ctx context.Context, conf *PrometheusServiceMeterConfig, name string) *prompb.MetricFamily {
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

	return mfs[name]
}

func matchBucket(conf *PrometheusServiceMeterConfig, values []float64) types.GomegaMatcher {
	matcher := make([]types.GomegaMatcher, 0, len(conf.HistogramBoundaries))

	for _, bound := range conf.HistogramBoundaries {
		count := uint64(0)
		for _, value := range values {
			if value <= bound {
				count++
			}
		}

		matcher = append(matcher, PointTo(MatchFields(IgnoreExtras, Fields{
			"CumulativeCount": PointTo(Equal(count)),
			"UpperBound":      PointTo(Equal(bound)),
		})))
	}

	return ContainElements(matcher)
}
