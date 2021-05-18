package propagator_test

import (
	"testing"

	"github.com/SpazioDati/go-utils/propagator"
	"github.com/stretchr/testify/assert"
)

func TestPropagatorConst(t *testing.T) {
	assert.Equal(t, "X-Dl-Request-Id", propagator.HDRSDRequestID)
}

func TestPropagatorBase(t *testing.T) {
	pr := propagator.Propagator{}

	pr.Set("foo", "bar")
	assert.Equal(t, pr.Get("foo"), "bar")
}

func TestPropagatorAdvanced(t *testing.T) {
	assert := assert.New(t)

	// https://github.com/aws/aws-xray-sdk-python/blob/master/aws_xray_sdk/core/models/http.py
	// https://github.com/aws/aws-xray-sdk-python/blob/master/aws_xray_sdk/ext/util.py#L20
	// https://docs.aws.amazon.com/xray/latest/devguide/xray-concepts.html#xray-concepts-tracingheader
	// https://scanzia.spaziodati.eu/atoka/atoka-revenge/-/blob/develop/atoka/helpers/__init__.py

	for _, test := range []struct {
		id    string
		reqID string
	}{
		{
			"00-6040dce1ae43ffe2332af577aa0af6af-f060f1fc34bcb745-01",
			"1-6040dce1-ae43ffe2332af577aa0af6af",
		},
		{
			"00-6040dce1ae43ffe2332af577aa0af6af-f060f1fc34bcb745-00",
			"1-6040dce1-ae43ffe2332af577aa0af6af",
		},
	} {
		for _, key := range []string{"Traceparent", "traceparent"} {
			pr := propagator.Propagator{}
			pr.Set(key, test.id)
			assert.Equal(pr.Get(key), test.id)

			// and X-Dl-Request-Id
			assert.Equal(pr.Get(propagator.HDRSDRequestID), test.reqID)

			assert.Equal(len(pr.Keys()), 2)
		}
	}
}
