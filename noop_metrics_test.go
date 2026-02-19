package common_test

import (
	"testing"

	common "github.com/slackmgr/slack-manager-common"
)

func TestNoopMetrics(t *testing.T) {
	t.Parallel()

	// Ensure NoopMetrics implements the Metrics interface
	var m common.Metrics = &common.NoopMetrics{}

	// Ensure methods can be called without errors or panics
	m.RegisterCounter("", "")
	m.RegisterGauge("", "")
	m.RegisterHistogram("", "", []float64{})
	m.Add("", 0)
	m.Inc("", "")
	m.Set("", 0)
	m.Observe("", 0)
}
