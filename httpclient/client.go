package httpclient

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

var traceEnabled = false

func New(opts Options) *resty.Client {
	return resty.
		NewWithClient(opts.HTTPClient).
		SetLogger(opts.Logger).
		SetTimeout(opts.Timeout).
		SetRetryCount(opts.Retries).
		SetHeader("User-Agent", opts.UserAgent).
		AddRetryCondition(RetryCondition()).
		OnBeforeRequest(OnBeforeRequest(opts.Logger)).
		OnAfterResponse(OnAfterResponse(opts.Logger)).
		OnError(OnError(opts.Logger))
}

func EnableTrace(timeout time.Duration) {
	traceEnabled = true
	go func(timeout time.Duration) {
		time.Sleep(timeout)
		traceEnabled = false
	}(timeout)
}

func IsTraceEnabled() bool {
	return traceEnabled
}

func OnBeforeRequest(logger resty.Logger) resty.RequestMiddleware {
	return func(c *resty.Client, r *resty.Request) error {
		if !IsTraceEnabled() {
			return nil
		}

		r.EnableTrace()
		return nil
	}
}

func OnAfterResponse(logger resty.Logger) resty.ResponseMiddleware {
	return func(c *resty.Client, r *resty.Response) error {
		return Doer(logger, r.Request.TraceInfo())
	}
}

func OnError(logger resty.Logger) resty.ErrorHook {
	return func(r *resty.Request, err error) {
		_ = Doer(logger, r.TraceInfo())
	}
}

func Doer(logger resty.Logger, ti resty.TraceInfo) (err error) {
	if !IsTraceEnabled() {
		return
	}

	hash := map[string]interface{}{
		"DNSLookup":      ti.DNSLookup,
		"ConnTime":       ti.ConnTime,
		"TCPConnTime":    ti.TCPConnTime,
		"TLSHandshake":   ti.TLSHandshake,
		"ServerTime":     ti.ServerTime,
		"ResponseTime":   ti.ResponseTime,
		"TotalTime":      ti.TotalTime,
		"IsConnReused":   ti.IsConnReused,
		"IsConnWasIdle":  ti.IsConnWasIdle,
		"ConnIdleTime":   ti.ConnIdleTime,
		"RequestAttempt": ti.RequestAttempt,
	}
	if ti.RemoteAddr != nil {
		hash["RemoteAddr"] = ti.RemoteAddr.String()
	}
	logger.Warnf("Resty TraceInfo: %v", hash)

	return
}

func RetryCondition() resty.RetryConditionFunc {
	return func(r *resty.Response, err error) bool {
		return err != nil || r.StatusCode() >= http.StatusInternalServerError
	}
}

func TraceEnablerMW(logger resty.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		to := 15 * time.Minute
		if timeout, ok := c.GetQuery("to"); ok {
			if converted, err := strconv.Atoi(timeout); err == nil {
				to = time.Duration(converted) * time.Minute
			}
		}
		logger.Warnf("Enabling resty trace for %d minutes", int(to.Minutes()))
		EnableTrace(to)
		c.String(http.StatusOK, "OK")
	}
}
