package iopipe

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// LoggerPluginConfig is the logger plugin configuration
type LoggerPluginConfig struct {
	Key string
}

type loggerPlugin struct {
	LoggerPluginConfig
}

func (p *loggerPlugin) Meta() *PluginMeta {
	return &PluginMeta{
		Name:     "@iopipe/logger",
		Version:  "0.1.0",
		Homepage: "https://github.com/iopipe/iopipe-go#logger-plugin",
		Enabled:  p.Enabled(),
	}
}

func (p *loggerPlugin) Enabled() bool {
	return true
}

func (p *loggerPlugin) PreSetup(agent *Agent) {
	log.SetFormatter(JSONFormatter{})
	agent.log = NewLogger()
}

func (p *loggerPlugin) PostSetup(agent *Agent)                              {}
func (p *loggerPlugin) PreInvoke(ctx context.Context, payload interface{})  {}
func (p *loggerPlugin) PostInvoke(ctx context.Context, payload interface{}) {}
func (p *loggerPlugin) PreReport(report *Report)                            {}
func (p *loggerPlugin) PostReport(report *Report)                           {}

// LoggerPlugin loads the logger plugin
func LoggerPlugin(config LoggerPluginConfig) PluginInstantiator {
	return func() Plugin {
		return &loggerPlugin{config}
	}
}

// JSONEntry is a log message
type JSONEntry struct {
	Timestamp string `json:"timestamp"`
	Name      string `json:"name"`
	Severity  string `json"severity"`
	Message   string `json:"message"`
}

// JSONFormatter formats logs into JSON
type JSONFormatter struct{}

func (f JSONFormatter) Format(entry *log.Entry) ([]byte, error) {
	jsonEntry := JSONEntry{
		Timestamp: entry.Time.UTC().Format("2006-01-02 15:04:05.000"),
		Name:      "root",
		Severity:  entry.Level.String(),
		Message:   entry.Message,
	}

	serialized, err := json.Marshal(jsonEntry)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}

	return append(serialized, '\n'), nil
}
