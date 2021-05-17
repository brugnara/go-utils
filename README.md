# go-utils
Package go-utils provides some functions useful when dealing with distributed tracing and metrics. It uses OpenTelemetry for both of the features.

## Install
```go get github.com/SpazioDati/go-utils```

### Using Go Modules
You can just import the package:

```import github.com/SpazioDati/go-utils```

## Overview
To use this package, you first need to initialize OpenTelemetry. Example:
```
func initOtel() func() {
    // Setup a sample ratio based on your needs
    sampleRatio := 0.5
    
    return opentelemetry.Init(&opentelemetry.Options{
        Name: "your-service-name",
        Endpoint: OpenTelemetryEndpoint,
        Sampler: opentelemetry.SampleByRatio(sampleRatio),
        Attributes: map[string]string{
            "service.name": "your-service-name",
            "service.environment": "current-environment"
        }
    })
}
...
func main() {
    cleanup := initOtel()
    defer cleanup()
}
```

Now you will have access to:
- a tracer initialized with the provided options via `opentelemetry.GetTracer()`;
- a meter initialized with the provided options via `global.Meter("meter-name")`.

To export traces and metrics from the OpenTelemetry SDK to the OpenTelemetry Collector you can check [this example](https://github.com/open-telemetry/opentelemetry-go/tree/main/example/otel-collector).

In particular, in order to export the metrics you can use [Prometheus Exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/prometheusexporter).


