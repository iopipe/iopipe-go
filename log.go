package iopipe

import (
	log "github.com/sirupsen/logrus"
)

var logger *log.Logger

func init() {
	logger = log.New()

	logger.SetLevel(log.InfoLevel)
}

func enableDebugMode() {
	logger.SetLevel(log.DebugLevel)
}
