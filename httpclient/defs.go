package httpclient

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	userAgent = "GoFeeFee/1.0"
)

type Options struct {
	HTTPClient *http.Client
	Logger     resty.Logger
	Timeout    time.Duration
	Retries    int
}
