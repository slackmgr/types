package types_test

import (
	"testing"

	"github.com/slackmgr/types"
)

func TestNoopLogger(t *testing.T) {
	t.Parallel()

	// Ensure NoopLogger implements the Logger interface
	var m types.Logger = &types.NoopLogger{}

	// Ensure methods can be called without errors or panics
	m.Debug("")
	m.Debugf("%v", nil)
	m.Info("")
	m.Infof("%v", nil)
	m.Error("")
	m.Errorf("%v", nil)
	m.WithField("", nil)
	m.WithFields(nil)
}
