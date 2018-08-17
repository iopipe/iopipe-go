package iopipe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoggerPlugin_LoggerPlugin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		signerResponse := &SignerResponse{
			JWTAccess:     "foobar",
			SignedRequest: ts.URL,
			URL:           "https://some-url",
		}
		signerResponseJSONBytes, _ := json.Marshal(signerResponse)
		fmt.Fprintln(res, string(signerResponseJSONBytes))
	}))
	defer ts2.Close()

	oldRegion := os.Getenv("AWS_REGION")
	defer os.Setenv("AWS_REGION", oldRegion)
	os.Setenv("AWS_REGION", "mock")
	os.Setenv("MOCK_SERVER", ts2.URL)

	Convey("Logger plugin should be initialized by agent", t, func() {
		a := NewAgent(Config{
			Debug: True(),
			Plugins: []PluginInstantiator{
				LoggerPlugin(LoggerPluginConfig{}),
			},
		})

		So(len(a.plugins), ShouldEqual, 1)

		a.Reporter = func(report *Report) error {
			return nil
		}

		hw := NewHandlerWrapper(func(ctx context.Context, payload interface{}) (interface{}, error) {
			context, _ := FromContext(ctx)
			context.IOpipe.Log.Debug("Some debug message")
			context.IOpipe.Log.Info("Some info message")
			context.IOpipe.Log.Warning("Some warning message")
			context.IOpipe.Log.Error("Some error message")
			return nil, nil
		}, a)

		Convey("Logger plugin invoke hooks fired", func() {
			ctx := context.Background()
			hw.Invoke(ctx, nil)

			Convey("A logger label is added to report", func() {
				_, exists := hw.report.labels["@iopipe/plugin-logger"]
				So(exists, ShouldBeTrue)
			})
		})
	})
}
