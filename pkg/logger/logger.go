package logger

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Constants declaration section -----------------------------------------------
const (
	showError = "ERROR"
	showWarn  = "WARN"
	showInfo  = "INFO"
	showDebug = "DEBUG"
)

// SetupLogger setup and configure logging
func SetupLogger(logLevel string) (logger log.Logger) {

	logger = log.NewLogfmtLogger(os.Stdout)

	switch logLevel {
	case showError:
		logger = level.NewFilter(logger, level.AllowError())
	case showWarn:
		logger = level.NewFilter(logger, level.AllowWarn())
	case showDebug:
		logger = level.NewFilter(logger, level.AllowDebug())
	default:
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	return
}
