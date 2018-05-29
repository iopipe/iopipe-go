package iopipe

import (
	"testing"
)

type Region struct {
	Region string
	URL    string
}

func TestGetBaseURL(t *testing.T) {
	regions := []Region{
		Region{Region: "us-east-1", URL: "https://metrics-api.iopipe.com/"},
		Region{Region: "us-west-2", URL: "https://metrics-api.us-west-2.iopipe.com/"},
		// Fill out with thingies!
	}

	for _, region := range regions {
		if getBaseURL(region.Region) != region.URL {
			t.Errorf("got %q, expected %q", getBaseURL(region.Region), region.URL)
		}
	}
}
