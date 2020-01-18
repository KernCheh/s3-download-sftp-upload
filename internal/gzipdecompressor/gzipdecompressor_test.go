package gzipdecompressor

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sephora-sea/s3-download-sftp-upload/internal/s3helper"
)

func TestDecompressByteStream(t *testing.T) {
	type args struct {
		Bucket   string
		Key      string
		Filename string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "non gzip file",
			args: args{
				Bucket:   "bv-test-sftp",
				Key:      "product_feed/bazaar_voice/thailand.xml",
				Filename: "thailand.testout.xml",
			},
		},
		{
			name: "gzipped file",
			args: args{
				Bucket:   "luxola-assets-staging-aws-th",
				Key:      "product_feed/bazaar_voice/thailand.xml",
				Filename: "thailandzip.testout.xml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3ObjectInput := &s3.GetObjectInput{
				Bucket: aws.String(tt.args.Bucket),
				Key:    aws.String(tt.args.Key),
			}

			sess := session.Must(session.NewSession(&aws.Config{
				Region: aws.String("ap-southeast-1"),
			}))
			s3Svc := s3.New(sess)

			downloader := s3manager.NewDownloaderWithClient(s3Svc)
			downloader.Concurrency = 1 // Concurrency must be set to 1 for buffer to work properly

			pr, pw := io.Pipe()

			f, err := os.Create(tt.args.Filename)
			if err != nil {
				panic(err)
			}

			fmt.Printf("[S3 Helper] Commencing download of file from %s/%s\n", *s3ObjectInput.Bucket, *s3ObjectInput.Key)

			go func() {
				_, err := downloader.Download(&s3helper.WriteAtBuffer{Writer: pw}, s3ObjectInput)
				if err != nil {
					panic(fmt.Errorf("[S3 Helper] Error downloading file at %s/%s: %s", *s3ObjectInput.Bucket, *s3ObjectInput.Key, err.Error()))
				}
				pw.Close()
			}()

			var pr2 io.ReadCloser

			_, cancelFunc := context.WithCancel(context.Background())
			pr2, err = DecompressByteStream(pr, cancelFunc)
			if err != nil {
				panic(err)
			}

			bytesWritten2, err := io.Copy(f, pr2)

			f.Close()
			if err != nil {
				panic(err)
			}

			pr2.Close()
			fmt.Println(bytesWritten2, "bytes written")
		})
	}
}
