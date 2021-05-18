package opentelemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"

	"github.com/SpazioDati/go-utils/propagator"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	fcmScope = "https://www.googleapis.com/auth/firebase.messaging"
)

var (
	tracer        = otel.GetTracerProvider().Tracer("not-initialized")
	isInitialized = false
)

func GetTracer() oteltrace.Tracer {
	return tracer
}

// Init see: https://opentelemetry.io/docs/go/getting-started/
// If needed, start a new span: opentelemetry.GetTracer().Start(parentCtx, name, ..)
// See examples
func Init(options *Options) (cleanup func()) {
	ctx := context.Background()
	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(options.Endpoint),
	)
	exp, err := otlp.NewExporter(ctx, driver)
	if err != nil {
		panic(fmt.Sprintf("Failed to create the collector exporter: %v", err))
	}

	// https://aws-otel.github.io/docs/getting-started/go-sdk
	// https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/awsxrayexporter
	idg := xray.NewIDGenerator()

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			options.GetAttributes()...,
		),
	)
	if err != nil {
		panic(fmt.Sprintf("Could not set resources: %v", err))
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(
			sdktrace.Config{
				// DON'T use: sdktrace.TraceIDRatioBased(samplingRatio),
				DefaultSampler: options.Sampler,
			},
		),
		sdktrace.WithBatcher(
			exp,
			// add following two options to ensure flush
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(10),
		),
		sdktrace.WithIDGenerator(idg),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	tracer = otel.GetTracerProvider().Tracer(options.Name)
	SetInitialized(true)

	// Setup metrics
	cont := controller.New(
		processor.New(
			simple.NewWithHistogramDistribution(),
			exp,
		),
		controller.WithCollectPeriod(5*time.Second),
		controller.WithPusher(exp),
	)
	global.SetMeterProvider(cont.MeterProvider())

	if err := cont.Start(context.Background()); err != nil {
		panic(fmt.Sprintf("Could not start metrics controller: %v", err))
	}

	return func() {
		// cleanup
		defer func() {
			err := cont.Stop(context.Background())
			if err != nil {
				otel.Handle(err)
			}
		}()
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			if err := exp.Shutdown(ctx); err != nil {
				otel.Handle(err)
			}
		}()
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				otel.Handle(err)
			}
		}()
	}
}

func SetInitialized(status bool) {
	isInitialized = status
}

func GetHTTPClient() *http.Client {
	if isInitialized {
		return &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	}

	return &http.Client{}
}

func GinMW() func(c *gin.Context) {
	return func(c *gin.Context) {
		// set headers for client
		headers := propagator.Propagator{}
		propagation.TraceContext{}.Inject(c.Request.Context(), headers)
		xray.Propagator{}.Inject(c.Request.Context(), headers)

		for k, v := range headers {
			c.Header(k, v)
		}
	}
}

// GetFCMHTTPClient returns an HttpClient able to performs requests setting a valid Authorization header
func GetFCMHTTPClient(googleCredentials []byte) *http.Client {
	if isInitialized {
		return &http.Client{
			Transport: otelhttp.NewTransport(CustomFCMTransport(nil, googleCredentials)),
		}
	}

	return &http.Client{}
}

// GetTracingHeaders returns tracing headers computed from the given context
func GetTracingHeaders(ctx context.Context, fromHeaders map[string]string) (headers map[string]string) {
	if fromHeaders != nil {
		headers = fromHeaders
	} else {
		headers = map[string]string{}
	}
	// not testable but makes sense:
	if ctx == nil {
		return
	}
	//
	tmp := propagator.Propagator{}
	propagation.TraceContext{}.Inject(ctx, tmp)
	xray.Propagator{}.Inject(ctx, tmp)

	for k, v := range tmp {
		headers[k] = v
	}

	return
}

func DecorateLogger(ctx context.Context, logger *zap.Logger) *zap.Logger {
	params := []zap.Field{}
	for k, v := range GetTracingHeaders(ctx, nil) {
		params = append(params, zap.String(k, v))
	}
	return logger.With(params...)
}
