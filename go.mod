module github.com/SpazioDati/go-utils

go 1.16

require (
	github.com/gin-gonic/gin v1.7.1
	github.com/go-resty/resty/v2 v2.6.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.18.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.18.0
	go.opentelemetry.io/contrib/propagators/aws v0.18.0
	go.opentelemetry.io/otel v0.18.0
	go.opentelemetry.io/otel/exporters/otlp v0.18.0
	go.opentelemetry.io/otel/metric v0.18.0 // indirect
	go.opentelemetry.io/otel/sdk v0.18.0
	go.opentelemetry.io/otel/sdk/metric v0.18.0 // indirect
	go.opentelemetry.io/otel/trace v0.18.0
	go.uber.org/zap v1.16.0
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
