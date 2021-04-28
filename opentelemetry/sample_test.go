package opentelemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func params() sdktrace.SamplingParameters {
	return sdktrace.SamplingParameters{
		ParentContext: oteltrace.SpanContext{TraceState: oteltrace.TraceState{}},
	}
}

func TestNew(t *testing.T) {
	assert := assert.New(t)

	for _, test := range []struct {
		in  float64
		out interface{}
	}{
		{0, sdktrace.Drop},
		{-10, sdktrace.Drop},
		{1, sdktrace.RecordAndSample},
		{11, sdktrace.RecordAndSample},
	} {
		sampling := SampleByRatio(test.in)
		assert.NotNil(sampling)
		p := params()

		out := sampling.ShouldSample(p)

		assert.Equal(test.out, out.Decision)
		assert.Equal(p.ParentContext.TraceState, out.Tracestate)
		assert.NotEmpty(sampling.Description())
	}
}

func TestValidParentContext(t *testing.T) {
	assert := assert.New(t)
	p := params()

	stringed := []byte("adc76b00323e202d")
	traceID := [16]byte{}
	copy(traceID[:], stringed)

	stringed = []byte("12345678")
	spanID := [8]byte{}
	copy(spanID[:], stringed)

	p.ParentContext = oteltrace.SpanContext{
		TraceID:    oteltrace.TraceID(traceID),
		SpanID:     oteltrace.SpanID(spanID),
		TraceFlags: 0x0,
		TraceState: oteltrace.TraceState{},
	}

	// should drop because of the TraceFlags given before!
	out := SampleByRatio(1).ShouldSample(p)
	assert.Equal(sdktrace.Drop, out.Decision)

	// this should sample because the parent is sampled
	p.ParentContext.TraceFlags = oteltrace.FlagsSampled
	out = SampleByRatio(0).ShouldSample(p)
	assert.Equal(sdktrace.RecordAndSample, out.Decision)
}

func TestRatio(t *testing.T) {
	assert := assert.New(t)

	for _, test := range []float64{.1, .2, .3, .4, .5, .6, .7, .8, .9} {
		sampling := SampleByRatio(test)
		assert.NotNil(sampling)
		p := params()

		dropped := false
		accepted := false
		for i := 0; i < 100; i++ {
			out := sampling.ShouldSample(p)
			switch out.Decision {
			case sdktrace.RecordAndSample:
				accepted = true
			case sdktrace.Drop:
				dropped = true
			}
		}

		assert.True(dropped)
		assert.True(accepted)
	}
}
