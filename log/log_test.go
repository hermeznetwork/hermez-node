package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	Info("Test log.Infow", "value", 10)
	Infof("Test log.Infof %d", 10)
	Debugf("Test log.Debugf %d", 10)
	Error("Test log.Error", "value", 10)
	Errorf("Test log.Errorf %d", 10)
	Warnf("Test log.Warnf %d", 10)
}
