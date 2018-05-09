package iopipe

import (
	"fmt"
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"io/ioutil"
	"os"
)

//TODO: WIP for reporting to IOPipe
func ReportToIOPipe(report *Report) error {
	var (
		err error
		ipAddress = fmt.Sprintf("metrics-api.%s.iopipe.com", os.Getenv("AWS_REGION"))
		urlPath = "v0/event"
		networkTimeout = 1 * time.Second
	)

	tr := &http.Transport{
		DisableKeepAlives:      false,
		MaxIdleConns:           1, // TODO: is this equivalent to the maxCachedSessions in the js implementation
	}

	httpsClient := http.Client{Transport: tr, Timeout: networkTimeout}

	reportJSONBytes, _ := json.MarshalIndent(report, "", "  ")

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
