package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sephora-sea/s3-download-sftp-upload/internal/clock"
	"github.com/sephora-sea/s3-download-sftp-upload/internal/config"
	"github.com/sephora-sea/s3-download-sftp-upload/internal/s3helper"
	"github.com/sephora-sea/s3-download-sftp-upload/internal/sftphelper"
)

// Handler for lambda entrypoint
func Handler(ctx context.Context, s3Event events.S3Event) {
	var err error

	for _, record := range s3Event.Records {
		s3Entity := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3Entity.Bucket.Name, s3Entity.Object.Key)

		authorizedConfigID := config.GetInstance().AuthorizedConfigID
		if authorizedConfigID != "" && s3Entity.ConfigurationID != authorizedConfigID {
			fmt.Println("[Main] Unauthorized Configuration ID:", s3Entity.ConfigurationID)
			continue
		}

		s3ObjectInput := s3.GetObjectInput{
			Bucket: aws.String(s3Entity.Bucket.Name),
			Key:    aws.String(s3Entity.Object.Key),
		}

		if err = DownloadFromS3UploadToSftp(&s3ObjectInput); err != nil {
			panic(err)
		}
	}
}

func main() {
	lambda.Start(Handler)
}

// DownloadFromS3UploadToSftp downloads the file specified in s3ObjectInput and upload to SFTP server specified in environment variables `SFTP_HOST`, `SFTP_PORT`, `SFTP_USERNAME`, `SFTP_PASSWORD` and `UPLOAD_PATH`
func DownloadFromS3UploadToSftp(s3ObjectInput *s3.GetObjectInput) error {
	var wg sync.WaitGroup
	var errUploader error

	c, err := sftphelper.GetClient()
	if err != nil {
		return err
	}

	wg.Add(2)

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close() // pipewriter must be closed immediately or reader will not get the EOF signal

		defer func() {
			if err := recover(); err != nil {
				pw.Close()
				fmt.Println("Error in goroutine:", err)
				wg.Done()
				panic(err)
			}
		}()

		err = s3helper.DownloadToMemoryBuffer(s3ObjectInput, pw)
		if err != nil {
			panic(err)
		}

		wg.Done()
	}()

	go func() {
		defer pr.Close()

		errUploader = c.Upload(pr, config.GetInstance().UploadPath, GetFileName(&clock.RealClock{}, *s3ObjectInput.Key))
		if err != nil {
			fmt.Println("Error in goroutine:", errUploader.Error())
		}

		wg.Done()
	}()

	wg.Wait()

	return errUploader
}

// GetFileName gives the filename of the s3 key prefixed with a timestamp in the format 20060102150405filename.ext, referencing from Mon Jan 2 15:04:05 -0700 MST 2006
func GetFileName(clock clock.Clock, s3key string) string {
	ss := strings.Split(s3key, "/")
	return clock.Now().Format("20060102150405") + ss[len(ss)-1]
}
