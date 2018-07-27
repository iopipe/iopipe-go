package iopipe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// LoggerPluginConfig is the logger plugin configuration
type LoggerPluginConfig struct{}

type loggerPlugin struct {
	LoggerPluginConfig
	proxyWriter *ProxyWriter
	uploads     []string
}

func (p *loggerPlugin) Meta() *PluginMeta {
	return &PluginMeta{
		Name:     "@iopipe/logger",
		Version:  "0.1.0",
		Homepage: "https://github.com/iopipe/iopipe-go#logger-plugin",
		Enabled:  p.Enabled(),
		Uploads:  p.uploads,
	}
}

func (p *loggerPlugin) Enabled() bool {
	return true
}

func (p *loggerPlugin) PreSetup(agent *Agent) {
	if agent.log == nil {
		agent.log = NewLogger()
	}

	agent.log.Formatter = JSONFormatter{}
	agent.log.Out = p.proxyWriter
}

func (p *loggerPlugin) PostSetup(agent *Agent)                              {}
func (p *loggerPlugin) PreInvoke(ctx context.Context, payload interface{})  {}
func (p *loggerPlugin) PostInvoke(ctx context.Context, payload interface{}) {}

func (p *loggerPlugin) PreReport(report *Report) {
	defer p.proxyWriter.Reset()

	var (
		err            error
		networkTimeout = 1 * time.Second
	)

	signedRequest, err := GetSignedRequest(report, ".log")
	if err != nil {
		report.agent.log.Error(err)
		return
	}

	tr := &http.Transport{
		DisableKeepAlives: false,
		MaxIdleConns:      1, // is this equivalent to the maxCachedSessions in the js implementation
	}
	httpsClient := http.Client{Transport: tr, Timeout: networkTimeout}

	req, err := http.NewRequest("PUT", signedRequest.SignedRequest, p.proxyWriter)
	if err != nil {
		report.agent.log.Error(err)
		return
	}

	res, err := httpsClient.Do(req)
	if err != nil {
		report.agent.log.Error(err)
		return
	}

	report.agent.log.Debug("Log Data Upload Status: ", res.StatusCode)

	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	report.agent.log.Debug("Log Data Upload Response: ", string(bodyBytes))

	p.uploads = append(p.uploads, signedRequest.JWTAccess)
}

func (p *loggerPlugin) PostReport(report *Report) {}

// LoggerPlugin loads the logger plugin
func LoggerPlugin(config LoggerPluginConfig) PluginInstantiator {
	return func() Plugin {
		return &loggerPlugin{
			LoggerPluginConfig: config,
			proxyWriter:        NewProxyWriter(),
		}
	}
}

// JSONEntry is a JSON log message
type JSONEntry struct {
	Timestamp string `json:"timestamp"`
	Name      string `json:"name"`
	Severity  string `json"severity"`
	Message   string `json:"message"`
}

// JSONFormatter formats logs into JSON
type JSONFormatter struct{}

// Format formats a log entry into JSON
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

// ProxyWriter proxies stderr and writes logs to buffer
type ProxyWriter struct {
	buffer   *bytes.Buffer
	proxyOut io.Writer
}

// NewProxyWriter returns a new proxy log writer
func NewProxyWriter() *ProxyWriter {
	return &ProxyWriter{
		buffer:   &bytes.Buffer{},
		proxyOut: os.Stderr,
	}
}

func (w *ProxyWriter) Read(p []byte) (int, error) {
	return w.buffer.Read(p)
}

// Reset resets the memory buffer
func (w *ProxyWriter) Reset() {
	w.buffer.Reset()
}

func (w *ProxyWriter) Write(p []byte) (int, error) {
	w.buffer.Write(p)

	return w.proxyOut.Write(p)
}
