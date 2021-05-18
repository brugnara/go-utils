package opentelemetry_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/SpazioDati/go-utils/opentelemetry"
	"github.com/SpazioDati/go-utils/propagator"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func headers() []string {
	return []string{propagator.HDRSDRequestID}
}

func ExampleInit() {
	cleanup := Init(&Options{
		Name:     "cool project name",
		Endpoint: "1.2.3.4:4317",
		Sampler:  SampleByRatio(1 / 1000),
		Attributes: map[string]string{
			"service.name":    fmt.Sprintf("%s - %s", "project", "unstable"),
			"service.version": "unstable",
		},
	})
	defer cleanup()
}

func ExampleGinMW() {
	r := gin.New()

	r.Use(otelgin.Middleware("foobar"))
	r.Use(GinMW())

	r.GET("/", func(c *gin.Context) {
		span := oteltrace.SpanFromContext(c.Request.Context())
		span.SetAttributes(
			attribute.String("foo", "foo"),
			attribute.String("bar", "bar"),
		)
		// ...
	})

	r.GET("/new-span", func(c *gin.Context) {
		ctx, span := GetTracer().Start(
			c.Request.Context(),
			"this span is a child of the one created by the mw",
			oteltrace.WithAttributes(
				attribute.String("searchID", "searchID"),
				attribute.String("slug", "bla bla"),
				attribute.String("lang", "bli"),
			),
		)
		defer span.End()

		span.SetAttributes(attribute.String("hasCookie", "false"))

		// use ctx when needed, for example:
		_ = GetTracingHeaders(ctx, nil)
		/*
			// then:
			resp, _ := resty.New().
				SetHeaders(opentelemetry.GetTracingHeaders(ctx, nil)).
				Get("http://atoka.io/foo-bar/")
		*/
	})
	//..
}

func TestInit(t *testing.T) {
	cleanup := Init(&Options{})
	assert.NotNil(t, cleanup)

	// if something is wrong inside the cleanup, the test will fail
	cleanup()
}

func TestMwEmptyTrace(t *testing.T) {
	assert := assert.New(t)

	for _, test := range headers() {
		r := gin.New()
		called := false

		r.GET("/", func(c *gin.Context) {
			called = true
			assert.Empty(c.Get(test))
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)

		assert.Empty(w.Header().Get(test))
		assert.True(called)
	}
}

func TestMwTraced(t *testing.T) {
	assert := assert.New(t)

	for _, test := range headers() {
		r := gin.New()
		called := false

		r.Use(otelgin.Middleware("foobar"))
		r.Use(GinMW())

		r.GET("/", func(c *gin.Context) {
			called = true
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)

		assert.NotEmpty(w.Header().Get(test))
		assert.True(called)
	}
}

func TestGetTraceHeaders(t *testing.T) {
	assert := assert.New(t)

	r := gin.New()
	called := false

	r.Use(otelgin.Middleware("foobar"))
	r.Use(GinMW())

	r.GET("/", func(c *gin.Context) {
		called = true
		c.String(http.StatusOK, "ok")

		for _, header := range headers() {
			out := GetTracingHeaders(c.Request.Context(), nil)
			assert.NotEmpty(out[header])
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.True(called)
}

func TestGetTraceHeadersWithMap(t *testing.T) {
	assert := assert.New(t)

	r := gin.New()
	called := false

	r.Use(GinMW())

	r.GET("/", func(c *gin.Context) {
		called = true
		c.String(http.StatusOK, "ok")
		hash := map[string]string{
			"foo": "bar",
			"baz": "bar",
			"biz": "baz",
			"bar": "foo",
		}

		outHeaders := GetTracingHeaders(c, hash)
		for k, v := range hash {
			assert.Equal(v, outHeaders[k])
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.True(called)
}

func TestNil(t *testing.T) {
	assert.NotPanics(t, func() {
		GetTracingHeaders(context.TODO(), nil)
	})
}

func TestDecorateLoggerEmpty(t *testing.T) {
	logger, _ := zap.NewProduction()
	assert.Equal(t, logger, DecorateLogger(context.TODO(), logger), "should not change since no headers are present")
}

func TestDecorateLoggerTricky(t *testing.T) {
	assert := assert.New(t)
	r := gin.New()
	called := false

	r.Use(otelgin.Middleware("foobar"))
	r.Use(GinMW())

	r.GET("/", func(c *gin.Context) {
		called = true
		c.String(http.StatusOK, "ok")

		logger, _ := zap.NewProduction()
		assert.NotEqual(logger, DecorateLogger(c.Request.Context(), logger), "should not change since no headers are present")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.True(called)
}

func TestGetHTTPClient(t *testing.T) {
	assert := assert.New(t)

	SetInitialized(true)
	a := GetHTTPClient()
	b := GetHTTPClient()
	assert.True(&a != &b)

	SetInitialized(false)
	a = GetHTTPClient()
	b = GetHTTPClient()
	assert.True(&a != &b)
}
