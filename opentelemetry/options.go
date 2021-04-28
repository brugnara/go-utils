package opentelemetry

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Options struct {
	Endpoint   string
	Name       string
	Sampler    trace.Sampler
	Attributes map[string]string
}

func (o Options) GetAttributes() []attribute.KeyValue {
	ret := []attribute.KeyValue{}
	for k, v := range o.Attributes {
		ret = append(ret, attribute.String(k, v))
	}
	return ret
}
