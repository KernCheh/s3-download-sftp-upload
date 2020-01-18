package s3helper

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// WriteAtBuffer is a workaround to implement the io.WriterAt interface on a io.Writer.
// Since WriteAt position cannot be supported in io.Writer,
// it is the user's responsibility to ensure WriteAt() method is called sequentially.
type WriteAtBuffer struct {
	Writer io.Writer
}

// WriteAt writes a slice of bytes to the end of the buffer.
// The number of bytes written will be returned, or error.
// The positioning argument is ignored.
func (fw *WriteAtBuffer) WriteAt(p []byte, _ int64) (n int, err error) {
	return fw.Writer.Write(p)
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

	if _, err = downloader.Download(&WriteAtBuffer{Writer: pw}, s3ObjectInput); err != nil {
		return fmt.Errorf("[S3 Helper] Error downloading file at %s/%s: %s", *s3ObjectInput.Bucket, *s3ObjectInput.Key, err.Error())
	}

	fmt.Printf("[S3 Helper] Download file complete: %s/%s\n", *s3ObjectInput.Bucket, *s3ObjectInput.Key)

	return nil
}

// DeleteObject deletes an s3 object
func DeleteObject(bucket, key string) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	}))
	s3Svc := s3.New(sess)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s3Svc.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println("[S3 Helper] Error:", aerr.Error())
				return aerr.OrigErr()
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return fmt.Errorf("[S3 Helper] Error deleting file from %s/%s: %s", bucket, key, err.Error())
		}
	}

	return nil
}
