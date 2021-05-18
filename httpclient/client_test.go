package httpclient_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	. "github.com/SpazioDati/go-utils/httpclient"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)
	client := New(Options{
		HTTPClient: &http.Client{},
		Logger:     nil,
		Timeout:    1 * time.Second,
		Retries:    1,
		UserAgent:  "foobar",
	})

	assert.Equal(client.Header.Get("User-Agent"), "foobar")
	assert.Equal(client.RetryCount, 1)
	assert.Equal(client.GetClient().Timeout, 1*time.Second)
}

func TestRetryCondition(t *testing.T) {
	assert := assert.New(t)
	fn := RetryCondition()

	assert.True(fn(nil, errors.New("foobar")), "true if an error occured")

	// checks for 500 >= httpStatusCode <= 513
	for code := http.StatusInternalServerError; code <= http.StatusNetworkAuthenticationRequired; code++ {
		response := &resty.Response{
			RawResponse: &http.Response{
				StatusCode: code,
			},
		}

		assert.True(fn(response, nil), "true if no error occured and StatusCode >= 500")
	}

	// checks for 101 >= httpStatusCode < 500
	for code := http.StatusContinue; code < http.StatusInternalServerError; code++ {
		response := &resty.Response{
			RawResponse: &http.Response{
				StatusCode: code,
			},
		}

		assert.False(fn(response, nil), "false if no error occured and StatusCode < 500")
		assert.True(fn(response, errors.New("foobar")), "true if an error occured even with StatusCoda < 500")
	}
}

func TestEnableTrace(t *testing.T) {
	assert := assert.New(t)

	assert.False(IsTraceEnabled())

	EnableTrace(0 * time.Millisecond)
	assert.True(IsTraceEnabled())
	done := make(chan bool)
	go func() {
		time.Sleep(1 * time.Millisecond)
		assert.False(IsTraceEnabled())
		done <- true
	}()
	assert.True(<-done)
}

func TestOnAfterResponse(t *testing.T) {
	assert := assert.New(t)
	mock := &logMock{}
	fn := OnAfterResponse(mock)

	assert.Nil(fn(&resty.Client{}, &resty.Response{
		Request: &resty.Request{},
	}))
	assert.Empty(mock)

	EnableTrace(0 * time.Millisecond)
	assert.Nil(fn(&resty.Client{}, &resty.Response{
		Request: &resty.Request{},
	}))
	assert.NotEmpty(mock)
	assert.Equal("warn", mock.Type)
}

func TestOnBeforeRequest(t *testing.T) {
	assert := assert.New(t)
	request := &resty.Request{}
	fn := OnBeforeRequest(&logMock{})

	assert.Nil(fn(&resty.Client{}, request))

	EnableTrace(0 * time.Millisecond)
	assert.Nil(fn(&resty.Client{}, request))
}

func TestTraceEnablerMW(t *testing.T) {
	for _, test := range []struct {
		in  string
		out string
	}{
		{"10", "10"},
		{"", "15"},
		{"foobar", "15"},
	} {
		r := gin.New()
		logger := &logMock{}

		r.GET("/", TraceEnablerMW(logger))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/?to=%s", test.in), nil)
		r.ServeHTTP(w, req)

		out := fmt.Sprintf(logger.Format, logger.Values...)
		match := regexp.MustCompile(`(\d+)`).FindStringSubmatch(out)
		assert.Equal(t, test.out, match[0])
	}
}
