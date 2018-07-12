package iopipe

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSigner_GetSignerURL(t *testing.T) {
	regions := []Region{
		Region{Region: "", URL: "https://signer.us-east-1.iopipe.com/"},
		Region{Region: "ap-northeast-1", URL: "https://signer.ap-northeast-1.iopipe.com/"},
		Region{Region: "ap-southeast-2", URL: "https://signer.ap-southeast-2.iopipe.com/"},
		Region{Region: "eu-west-1", URL: "https://signer.eu-west-1.iopipe.com/"},
		Region{Region: "us-east-1", URL: "https://signer.us-east-1.iopipe.com/"},
		Region{Region: "us-east-2", URL: "https://signer.us-east-2.iopipe.com/"},
		Region{Region: "us-west-1", URL: "https://signer.us-west-1.iopipe.com/"},
		Region{Region: "us-west-2", URL: "https://signer.us-west-2.iopipe.com/"},
	}

	Convey("GetSignerURL should return the correct URL for the region", t, func() {
		for _, region := range regions {
			So(GetSignerURL(region.Region), ShouldEqual, region.URL)
		}
	})
}
