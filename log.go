package iopipe

import (
	log "github.com/sirupsen/logrus"
)

func NewLogger() *log.Logger {
	logger := log.New()
	logger.SetLevel(log.InfoLevel)
	return logger
}
