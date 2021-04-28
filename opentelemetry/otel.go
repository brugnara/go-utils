package opentelemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucademenego99/go-utils/propagator"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
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

	return func() {
		// cleanup
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
