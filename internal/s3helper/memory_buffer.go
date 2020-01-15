package s3helper

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type writeAtBuffer struct {
	w io.Writer
}

// WriteAt writes a slice of bytes to a buffer starting at the position provided
// The number of bytes written will be returned, or error. Can overwrite previous
// written slices if the write ats overlap.
func (fw *writeAtBuffer) WriteAt(p []byte, pos int64) (n int, err error) {
	return fw.w.Write(p)
}

// DownloadToMemoryBuffer downloads a file from s3 to a io.PipeWriter. Closes the pipewriter afterwards
func DownloadToMemoryBuffer(s3ObjectInput *s3.GetObjectInput, pw *io.PipeWriter) error {
	var err error

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	}))
	s3Svc := s3.New(sess)
	downloader := s3manager.NewDownloaderWithClient(s3Svc)
	downloader.Concurrency = 1 // Concurrency must be set to 1 for buffer to work properly

	fmt.Printf("[S3 Helper] Commencing download of file from %s/%s\n", *s3ObjectInput.Bucket, *s3ObjectInput.Key)

	if _, err = downloader.Download(&writeAtBuffer{w: pw}, s3ObjectInput); err != nil {
		return fmt.Errorf("[S3 Helper] Error downloading file at %s/%s: %s", *s3ObjectInput.Bucket, *s3ObjectInput.Key, err.Error())
	}

	fmt.Printf("[S3 Helper] Download file complete: %s/%s\n", *s3ObjectInput.Bucket, *s3ObjectInput.Key)

	return nil
}
