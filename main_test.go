package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sephora-sea/s3-download-sftp-upload/internal/clock"
)

func TestgetFileName(t *testing.T) {

	type args struct {
		clock clock.Clock
		s3key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "file format test",
			args: args{
				clock: &clock.TestClock{},
				s3key: "product_feed/testing/default.xml",
			},
			want: "20060102150405default.xml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFileName(tt.args.clock, tt.args.s3key); got != tt.want {
				t.Errorf("getFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloadFromS3UploadToSftp(t *testing.T) {
	// Warning: This test will fail if env variables for sftp server is not set properly
	// defer profile.Start(profile.MemProfile).Stop()
	type args struct {
		ctx           context.Context
		s3ObjectInput *s3.GetObjectInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "try 40mb file",
			args: args{
				ctx: context.Background(),
				s3ObjectInput: &s3.GetObjectInput{
					Bucket: aws.String("bv-test-sftp"),
					Key:    aws.String("product_feed/bazaar_voice/default.xml"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DownloadFromS3UploadToSftp(tt.args.ctx, tt.args.s3ObjectInput); (err != nil) != tt.wantErr {
				t.Errorf("DownloadFromS3UploadToSftp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
