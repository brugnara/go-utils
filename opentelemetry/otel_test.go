package opentelemetry_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/lucademenego99/go-utils/opentelemetry"
	"github.com/lucademenego99/go-utils/propagator"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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
