package main

import (
	"context"
	"fmt"
	"io"
	"strings"

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

		if err = DownloadFromS3UploadToSftp(context.Background(), &s3ObjectInput); err != nil {
			panic(err)
		}
	}
}

// DownloadFromS3UploadToSftp downloads the file specified in s3ObjectInput and upload to SFTP server specified in environment variables `SFTP_HOST`, `SFTP_PORT`, `SFTP_USERNAME`, `SFTP_PASSWORD` and `UPLOAD_PATH`
func DownloadFromS3UploadToSftp(ctx context.Context, s3ObjectInput *s3.GetObjectInput) error {
	var errDownloader, errUploader error
	var c *sftphelper.Client

	chanUploaderOK := make(chan bool)

	pr, pw := io.Pipe()

	// We define a cancel context to ensure errors with the downloader will halt the uploader, and this function will exit gracefully with an error
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	go func() {
		defer pw.Close() // pipewriter must be closed immediately or reader will not get the EOF signal

		errDownloader = s3helper.DownloadToMemoryBuffer(s3ObjectInput, pw)
		if errDownloader != nil {
			fmt.Println("Error in downloader goroutine:", errDownloader.Error())
			cancelFunc()
		}
	}()

	go func() {
		defer pr.Close()

		if c == nil {
			c, errUploader = sftphelper.GetClient()
			if errUploader != nil {
				fmt.Println("Error in uploader goroutine:", errUploader.Error())
				cancelFunc()
				return
			}
		}

		// Uploader takes in a context to handle early cancellation. Please note that a corrupted file may exist in the remote SFTP server if the downloader terminates.
		errUploader = c.UploadWithContext(ctx, pr, config.GetInstance().UploadPath, getFileName(&clock.RealClock{}, *s3ObjectInput.Key))
		if errUploader != nil {
			fmt.Println("Error in uploader goroutine:", errUploader.Error())
			cancelFunc()
			return
		}

		chanUploaderOK <- true
	}()

	select {
	case <-chanUploaderOK:
		// No issue
	case <-ctx.Done():
		// cancelFunc called
		if errDownloader != nil {
			return errDownloader
		}
		if errUploader != nil {
			return errUploader
		}
		return ctx.Err()
	}

	return nil
}

// GetFileName gives the filename of the s3 key prefixed with a timestamp in the format 20060102150405filename.ext, referencing from Mon Jan 2 15:04:05 -0700 MST 2006
func getFileName(clock clock.Clock, s3key string) string {
	ss := strings.Split(s3key, "/")
	return clock.Now().Format("20060102150405") + ss[len(ss)-1]
}

func main() {
	lambda.Start(Handler)
}
