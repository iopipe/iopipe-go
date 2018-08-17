package iopipe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func getCollectorURL(region string) string {
	supportedRegions := map[string]struct{}{
		"ap-northeast-1": struct{}{},
		"ap-southeast-2": struct{}{},
		"eu-west-1":      struct{}{},
		"us-east-2":      struct{}{},
		"us-west-1":      struct{}{},
		"us-west-2":      struct{}{},
	}

	if region == "mock" {
		return os.Getenv("MOCK_SERVER")
	}

	if _, exists := supportedRegions[region]; exists {
		return fmt.Sprintf("https://metrics-api.%s.iopipe.com/v0/event", region)
	}

	return "https://metrics-api.iopipe.com/v0/event"
}

func sendReport(report *Report) error {
	var (
		err            error
		networkTimeout = 1 * time.Second
	)

	tr := &http.Transport{
		DisableKeepAlives: false,
		MaxIdleConns:      1, // is this equivalent to the maxCachedSessions in the js implementation
	}

	httpsClient := http.Client{Transport: tr, Timeout: networkTimeout}

	reportJSONBytes, _ := json.Marshal(report) //.MarshalIndent(report, "", "  ")
	report.agent.log.Debug("Sending report:\n", string(reportJSONBytes))

	uRL := getCollectorURL(os.Getenv("AWS_REGION"))
	req, err := http.NewRequest("POST", uRL, bytes.NewReader(reportJSONBytes))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := httpsClient.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	resbody, err := ioutil.ReadAll(res.Body)
	report.agent.log.Debug("IOpipe response: ", string(resbody))

	if err != nil {
		return err
	}

	return nil
}
