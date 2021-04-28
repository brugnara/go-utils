// useful because how propagation works: "go.opentelemetry.io/otel/propagation"
package propagator

import (
	"fmt"
	"strings"
)

// Propagator stores a map[string]string and implements https://pkg.go.dev/go.opentelemetry.io/otel@v0.17.0/propagation#TextMapCarrier
type Propagator map[string]string

// headers used across SD apps
const (
	HDRSDRequestID = "X-Dl-Request-Id"
)

// Get is the getter for key
func (p Propagator) Get(key string) string {
	return p[key]
}

// Set is the setter for key, value
func (p Propagator) Set(key, value string) {
	p[key] = value
	if key == "Traceparent" || key == "traceparent" {
		// https://www.w3.org/TR/trace-context/#traceparent-header
		// 00-6040dce1ae43ffe2332af577aa0af6af-f060f1fc34bcb745-00
		// 00: version
		// 6..f: trace-id
		// f..5: parent-id
		// 00: trace-flags
		tmp := strings.Split(value, "-")

		p[HDRSDRequestID] = fmt.Sprintf("1-%s-%s", tmp[1][:8], tmp[1][8:])
	}
}

func (p Propagator) Keys() []string {
	ret := make([]string, len(p))
	i := 0
	for k := range p {
		ret[i] = k
		i++
	}
	return ret
}
