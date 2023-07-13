package aws

import (
	"csi-api/app/env"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	AWS_S3_REGION = env.Config.S3.Region
	AWS_S3_BUCKET = env.Config.S3.S3BUCKET
	AWS_S3_KEY    = env.Config.S3.AccessKey
)

// Connect to AWS S3 bucket :
func ConnectAWS() (*session.Session, error) {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(AWS_S3_REGION),
		},
	)
	if err != nil {
		return sess, err
	}
	return sess, nil
}

// ListBuckets :
func ListBuckets() (resp *s3.ListBucketsOutput) {
	s3session := s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(AWS_S3_REGION),
	})))
	resp, err := s3session.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		panic(err)
	}
	return resp
}

// Upload file to bucket :
func PushS3Bucket(file *multipart.File, path string, filename string) (string, error) {
	sess, err := ConnectAWS()
	if err != nil {
		// Fail to connect to bucket
		return "", err
	}
	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET),   // Bucket to be used
		Key:    aws.String(path + filename), // Name of the file to be saved
		Body:   *file,                       // File
	})
	if err != nil {
		return "", err
	}

	return path + filename, nil
}

// Push doc to bucket :
func PushDocS3Bucket(file *os.File, path string, filename string) (string, error) {
	sess, err := ConnectAWS()
	if err != nil {
		// Fail to connect to bucket
		return "", err
	}
	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET),   // Bucket to be used
		Key:    aws.String(path + filename), // Name of the file to be saved
		Body:   file,                        // File
	})

	if err != nil {
		// Do your error handling here
		return "", err
	}

	return path + filename, nil
}

// Create signed Url :
func SignedURL(path string) (string, error) {
	var contactType string
	sess, err := ConnectAWS()
	if err != nil {
		// Fail to connect to bucket
		return "", err
	}
	// Create S3 service client
	svc := s3.New(sess)
	if strings.Contains(path, "icImg") {
		contactType = "image/jpeg"
	} else if strings.Contains(path, "csv") {
		contactType = "text/csv"
	} else if strings.Contains(path, "pdf") {
		contactType = "application/pdf"
	}

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket:              aws.String(AWS_S3_BUCKET),
		Key:                 aws.String(path),
		ResponseContentType: aws.String(contactType),
	})

	urlStr, err := req.Presign(10080 * time.Minute)
	if err != nil {
		return "", err
	}

	return urlStr, nil
}
