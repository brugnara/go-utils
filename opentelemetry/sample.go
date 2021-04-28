package opentelemetry

import (
	"fmt"
	"math/rand"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SampleByRatio implements Sampler interface: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#Sampler
// Be aware this will also check for an attribute named FlagSkip. If this is present, this will make the
// sampler skip too
type SampleByRatio float64

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (ratio SampleByRatio) ShouldSample(p sdktrace.SamplingParameters) (ret sdktrace.SamplingResult) {
	ret = sdktrace.SamplingResult{
		// https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#SamplingDecision
		Decision:   sdktrace.RecordAndSample,
		Tracestate: p.ParentContext.TraceState,
	}

	if p.ParentContext.IsValid() {
		if !p.ParentContext.IsSampled() {
			ret.Decision = sdktrace.Drop
		}
		return
	}

	if ratio >= 1 {
		return
	}

	if ratio <= 0 {
		ret.Decision = sdktrace.Drop
		return
	}

	// check with random
	if rnd := rand.Float64(); rnd > float64(ratio) {
		ret.Decision = sdktrace.Drop
	}

	return
}

func (ratio SampleByRatio) Description() string {
	return fmt.Sprintf("ratio: %f", ratio)
}
