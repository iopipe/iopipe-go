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

//TODO: WIP for reporting to IOPipe
func ReportToIOPipe(report *Report) error {
	var (
		err            error
		urlPath        = "v0/event"
		networkTimeout = 1 * time.Second
	)

	region := os.Getenv("AWS_REGION")
	// array of supported regions so we can easily look up
	// whether a region has its own collector
	// using empty structs takes up no space versus using, say, a bool
	supportedRegions := map[string]struct{}{
		"ap-northeast-1": struct{}{},
		"ap-southeast-2": struct{}{},
		"eu-west-1":      struct{}{},
		"us-east-2":      struct{}{},
		"us-west-1":      struct{}{},
		"us-west-2":      struct{}{},
	}

	ipAddress := "metrics-api.iopipe.com"
	if _, exists := supportedRegions[region]; exists {
		ipAddress = fmt.Sprintf("metrics-api.%s.iopipe.com", region)
	}

	tr := &http.Transport{
		DisableKeepAlives: false,
		MaxIdleConns:      1, // TODO: is this equivalent to the maxCachedSessions in the js implementation
	}

	httpsClient := http.Client{Transport: tr, Timeout: networkTimeout}

	reportJSONBytes, _ := json.MarshalIndent(report, "", "  ")

	// TODO defining 443 extraneous
	reqURL := fmt.Sprintf("https://%s:443/%s", ipAddress, urlPath)
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(reportJSONBytes))
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
	fmt.Println("body read from IOPIPE", string(resbody))
	if err != nil {
		return err
	}

	return nil
}
