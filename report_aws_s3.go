package iopipe

import (
	"os"
	"fmt"
	"bytes"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws"
)

var awsSharedSession *session.Session

func UploadFileToS3(fileName, fileContents string) {
	var err error

	// Create a single AWS session (we can re use this if we're uploading many files)
	if awsSharedSession == nil {
		awsSharedSession, err = session.NewSession()
		if err != nil {
			panic(err)
		}
	}

	svc := s3manager.NewUploader(awsSharedSession)

	fmt.Println("Starting AWS upload")

	// Config settings: this is where you choose the bucket, fileName etc.
	// of the file you're uploading.
	_, err = svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader([]byte(fileContents)),
		ACL:    aws.String("public-read"),
	})

	fmt.Println("Finished AWS upload")
	fmt.Println(
		fmt.Sprintf("Report available at https://s3.amazonaws.com/%s/%s",
			os.Getenv("S3_BUCKET"),
			fileName,
		),
	)

	if err != nil {
		panic(err)
	}
}
