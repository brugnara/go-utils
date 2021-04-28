package opentelemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(0, len(Options{}.GetAttributes()))

	opts := Options{
		Attributes: map[string]string{
			"foo": "bar",
			"baz": "biz",
		},
	}
	out := opts.GetAttributes()

	assert.Equal(2, len(out))

	for _, attr := range out {
		assert.NotEmpty(opts.Attributes[string(attr.Key)])
	}
}
