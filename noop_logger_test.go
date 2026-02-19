package common_test

import (
	"testing"

	common "github.com/slackmgr/slack-manager-common"
)

func TestNoopLogger(t *testing.T) {
	t.Parallel()

	// Ensure NoopLogger implements the Logger interface
	var m common.Logger = &common.NoopLogger{}

	// Ensure methods can be called without errors or panics
	m.Debug("")
	m.Debugf("", nil)
	m.Info("")
	m.Infof("", nil)
	m.Error("")
	m.Errorf("", nil)
	m.WithField("", nil)
	m.WithFields(nil)
}
