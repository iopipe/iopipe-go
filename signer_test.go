package iopipe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSigner_GetSignerURL(t *testing.T) {
	regions := []Region{
		Region{Region: "", URL: "https://signer.us-west-2.iopipe.com/"},
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

func TestSigner_GetSignedRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		signerResponse := &SignerResponse{
			JWTAccess:     "foobar",
			SignedRequest: "https://some-url",
			URL:           "https://some-url",
		}
		signerResponseJSONBytes, _ := json.Marshal(signerResponse)

		fmt.Fprintln(res, string(signerResponseJSONBytes))
	}))
	defer ts.Close()

	oldRegion := os.Getenv("AWS_REGION")
	defer os.Setenv("AWS_REGION", oldRegion)

	os.Setenv("AWS_REGION", "mock")
	os.Setenv("MOCK_SERVER", ts.URL)

	a := NewAgent(Config{})
	lc := &lambdacontext.LambdaContext{
		AwsRequestID:       "123",
		InvokedFunctionArn: "Foo::Bar::Baz",
	}
	hw := &HandlerWrapper{agent: a, lambdaContext: lc}
	r := NewReport(hw)

	Convey("GetSignedRequest should return a signed request", t, func() {
		res, err := GetSignedRequest(r, ".txt")

		So(err, ShouldBeNil)
		So(res.JWTAccess, ShouldEqual, "foobar")
		So(res.SignedRequest, ShouldEqual, "https://some-url")
		So(res.URL, ShouldEqual, "https://some-url")
	})
}
