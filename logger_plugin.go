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
	"sync"
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

func (p *loggerPlugin) PostSetup(agent *Agent) {}
func (p *loggerPlugin) PreInvoke(ctx context.Context, payload interface{}) {
	p.proxyWriter.Reset()
}

func (p *loggerPlugin) PostInvoke(ctx context.Context, payload interface{}) {
	context, _ := FromContext(ctx)

	if p.proxyWriter.Len() > 0 {
		context.IOpipe.Label("@iopipe/plugin-logger")
	}
}

func (p *loggerPlugin) PreReport(report *Report) {
	var (
		err            error
		networkTimeout = 60 * time.Second
	)

	if p.proxyWriter.Len() == 0 {
		report.agent.log.Debug("No log messages to upload, skipping")
		return
	}

	signedRequest, err := GetSignedRequest(report, ".log")
	if err != nil {
		report.agent.log.Debug(err)
		return
	}

	tr := &http.Transport{
		DisableKeepAlives: false,
		MaxIdleConns:      1, // is this equivalent to the maxCachedSessions in the js implementation
	}
	httpsClient := http.Client{Transport: tr, Timeout: networkTimeout}

	req, err := http.NewRequest("PUT", signedRequest.SignedRequest, p.proxyWriter)
	if err != nil {
		report.agent.log.Debug(err)
		return
	}

	res, err := httpsClient.Do(req)
	if err != nil {
		report.agent.log.Debug(err)
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
	mutex    sync.RWMutex
	proxyOut io.Writer
}

// NewProxyWriter returns a new proxy log writer
func NewProxyWriter() *ProxyWriter {
	return &ProxyWriter{
		buffer:   &bytes.Buffer{},
		proxyOut: os.Stderr,
	}
}

// Len returns the size of the memory buffer
func (w *ProxyWriter) Len() int {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.buffer.Len()
}

// Read reads from the memory buffer
func (w *ProxyWriter) Read(p []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.buffer.Read(p)
}

// Reset resets the memory buffer
func (w *ProxyWriter) Reset() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.buffer.Reset()
}

// Write writes bytes to the buffer
func (w *ProxyWriter) Write(p []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.buffer.Write(p)

	return w.proxyOut.Write(p)
}
