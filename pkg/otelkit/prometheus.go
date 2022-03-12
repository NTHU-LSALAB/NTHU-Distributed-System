package otelkit

import (
	"context"
	"net/http"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type PrometheusServiceMeterConfig struct {
	Addr                string    `long:"addr" env:"ADDR" description:"the prometheus exporter address" default:":2222"`
	Path                string    `long:"path" env:"PATH" description:"the prometheus exporter path" default:"/metrics"`
	Name                string    `long:"name" env:"NAME" description:"the unique instrumentation name" required:"true"`
	HistogramBoundaries []float64 `long:"histogram_boundaries" env:"HISTOGRAM_BOUNDARIES" env-delim:"," description:"the default histogram boundaries of prometheus" required:"true"`
}

// PrometheusServiceMeter provides 3 meters to measure:
// 1. Count number of requests
// 2. Measure response time
// // TODO: 3. Count number of error requests
type PrometheusServiceMeter struct {
	metric.Meter

	server                *http.Server
	requestCounter        syncint64.Counter
	requestErrorCounter   syncint64.Counter
	responseTimeHistogram syncint64.Histogram
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *PrometheusServiceMeter) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		attributes := []attribute.KeyValue{
			attribute.String("FullMethod", info.FullMethod),
		}

		// count request
		m.requestCounter.Add(ctx, 1, attributes...)

		start := time.Now()

		resp, err := handler(ctx, req)

		// error count request
		if err != nil {
			m.requestErrorCounter.Add(ctx, 1, attributes...)
		}

		// measure response time
		responseTime := time.Since(start)
		m.responseTimeHistogram.Record(ctx, responseTime.Milliseconds(), attributes...)

		return resp, err
	}
}

func (m *PrometheusServiceMeter) Close() error {
	return m.server.Close()
}

func NewPrometheusServiceMeter(ctx context.Context, conf *PrometheusServiceMeterConfig) *PrometheusServiceMeter {
	logger := logkit.FromContext(ctx).With(
		zap.String("path", conf.Path),
		zap.String("port", conf.Addr),
		zap.String("name", conf.Name),
	)

	exporter := newPrometheusExporter(conf, logger)
	server := newPrometheusServer(exporter, conf, logger)

	meter := exporter.MeterProvider().Meter(conf.Name)

	requestCounter, err := meter.SyncInt64().Counter("request", instrument.WithDescription("count number of requests"))
	if err != nil {
		logger.Fatal("failed to create requests counter", zap.Error(err))
	}

	requestErrorCounter, err := meter.SyncInt64().Counter("error_request", instrument.WithDescription("count number of error requests"))
	if err != nil {
		logger.Fatal("failed to create error requests counter", zap.Error(err))
	}

	responseTimeHistogram, err := meter.SyncInt64().Histogram("response_time", instrument.WithDescription("measure response time"))
	if err != nil {
		logger.Fatal("failed to create response time histogram", zap.Error(err))
	}

	return &PrometheusServiceMeter{
		server:                server,
		requestCounter:        requestCounter,
		requestErrorCounter:   requestErrorCounter,
		responseTimeHistogram: responseTimeHistogram,
	}
}

func newPrometheusExporter(conf *PrometheusServiceMeterConfig, logger *logkit.Logger) *prometheus.Exporter {
	config := prometheus.Config{
		DefaultHistogramBoundaries: conf.HistogramBoundaries,
	}

	c := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
	)

	exporter, err := prometheus.New(config, c)
	if err != nil {
		logger.Fatal("failed to create prometheus exporter", zap.Error(err))
	}

	return exporter
}

func newPrometheusServer(exporter *prometheus.Exporter, conf *PrometheusServiceMeterConfig, logger *logkit.Logger) *http.Server {
	server := &http.Server{
		Addr: conf.Addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == conf.Path {
				exporter.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		}),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error("failed to serve prometheus exporter", zap.Error(err))
		}
	}()

	logger.Info("serve prometheus exporter successfully")

	return server
}
