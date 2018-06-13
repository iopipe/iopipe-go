package iopipe

import (
	log "github.com/sirupsen/logrus"
)

var logger *log.Entry

func init() {
	log.SetLevel(log.InfoLevel)

	logger = log.WithFields(log.Fields{
		"name": "iopipe",
	})
}

func enableDebugMode() {
	log.SetLevel(log.DebugLevel)
}
