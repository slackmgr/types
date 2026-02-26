package types_test

import (
	"testing"

	"github.com/slackmgr/types"
)

func TestNoopMetrics(t *testing.T) {
	t.Parallel()

	// Ensure NoopMetrics implements the Metrics interface
	var m types.Metrics = &types.NoopMetrics{}

	// Ensure methods can be called without errors or panics
	m.RegisterCounter("", "")
	m.RegisterGauge("", "")
	m.RegisterHistogram("", "", []float64{})
	m.CounterAdd("", 0)
	m.CounterInc("", "")
	m.GaugeSet("", 0)
	m.GaugeAdd("", 0)
	m.Observe("", 0)
}
