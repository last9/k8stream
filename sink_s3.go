package main

import (
	"compress/gzip"
	"encoding/json"
	fmt "fmt"
	"io"
	"log"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Sink struct {
	Prefix  string `json:"prefix" validate:"required"`
	Region  string `json:"aws_region", validate:"required"`
	Bucket  string `json:"aws_bucket" validate:"required"`
	Profile string `json:"aws_profile" validate:"required"`
}

func (s *S3Sink) LoadConfig(b json.RawMessage) error {
	return loadConfig(b, s)
}

var s3s *session.Session
var s3Once sync.Once

func getSession(s *S3Sink) (*session.Session, error) {
	var err error
	s3Once.Do(func() {
		s3s, err = session.NewSession(&aws.Config{
			Region:      aws.String(s.Region),
			Credentials: credentials.NewSharedCredentials("", s.Profile),
		})
	})

	return s3s, err
}

func (s *S3Sink) Flush(uuid, filename string, d []byte) error {
	sess, err := getSession(s)
	if err != nil {
		return err
	}

	if sess == nil {
		return fmt.Errorf("Empty session. There was an error earlier")
	}

	reader, writer := io.Pipe()
	go func() {
		zw := gzip.NewWriter(writer)
		zw.Write(d)
		zw.Close()
		writer.Close()
	}()

	prefix := filepath.Join(s.Prefix, uuid)
	return uploadToS3(sess, s.Bucket, prefix, filename, reader)
}

func uploadToS3(
	s *session.Session,
	bucket, prefix, filename string,
	content io.Reader,
) error {
	uploader := s3manager.NewUploader(s)

	fname := fmt.Sprintf("%v/%v_k8s.proto.gz", prefix, filename)
	log.Println("Upload", fname)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fname),
		ACL:    aws.String("private"),
		Body:   content,
	})

	return err
}
