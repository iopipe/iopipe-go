package iopipe

import (
	"fmt"
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"io/ioutil"
)

//TODO: WIP for reporting to IOPipe
func ReportToIOPipe(report *Report) error {
	var (
		err error
		ipAddress string
		urlPath string
		networkTimeout time.Duration
	)

	tr := &http.Transport{
		DisableKeepAlives:      false,
		MaxIdleConns:           1, // TODO: is this equivalent to the maxCachedSessions in the js implementation
	}

	httpsClient := http.Client{Transport: tr, Timeout: networkTimeout}

	reportJSONBytes, _ := json.MarshalIndent(report, "", "  ")

	reqURL := fmt.Sprintf("%s:443/%s", ipAddress, urlPath)
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(reportJSONBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := httpsClient.Do(req)
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}
