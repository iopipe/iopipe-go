package iopipe

import (
	"github.com/sirupsen/logrus"
)

// NewLogger returns a new logger with default config
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	return logger
}
