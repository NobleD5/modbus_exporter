package logger

import (
	"fmt"
	"testing"
)

func TestSetupLogger(t *testing.T) {

	const (
		showError = "ERROR"
		showWarn  = "WARN"
		showInfo  = "INFO"
		showDebug = "DEBUG"
	)

	logLevels := []string{
		"",
		showError,
		showWarn,
		showInfo,
		showDebug,
	}

	for _, level := range logLevels {
		logger := SetupLogger(level)
		fmt.Println(logger)
	}

}
