package repo

import (
	"backstreetlinkv2/cmd/helper"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

var (
	NoObjectID = errors.New("object id is not found/empty")
)

type ObjectScanner struct {
	svc        s3iface.S3API
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
}

type ObjectConfig struct {
	AccessKey        string
	SecretKey        string
	Endpoint         string
	Region           string
	ForceS3PathStyle bool

	Bucket string
}

func NewObjectScanner(cfg ObjectConfig) (*ObjectScanner, error) {
	awsSession, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String(cfg.Region),
		S3ForcePathStyle: aws.Bool(cfg.ForceS3PathStyle),
	})

	if err != nil {
		return nil, err
	}

	s3Service := s3.New(awsSession)

	//if err := setSSEBucket(cfg.Bucket, s3Service); err != nil {
	//	return nil, err
	//}

	return &ObjectScanner{
		svc:        s3Service,
		uploader:   s3manager.NewUploaderWithClient(s3Service),
		downloader: s3manager.NewDownloaderWithClient(s3Service),
		bucket:     cfg.Bucket,
	}, nil
}

func setSSEBucket(bucketName string, service *s3.S3) error {
	defEnc := &s3.ServerSideEncryptionByDefault{
		SSEAlgorithm: aws.String(s3.ServerSideEncryptionAes256),
	}

	rule := &s3.ServerSideEncryptionRule{ApplyServerSideEncryptionByDefault: defEnc}
	rules := []*s3.ServerSideEncryptionRule{rule}
	serverConfig := &s3.ServerSideEncryptionConfiguration{Rules: rules}
	input := &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucketName), ServerSideEncryptionConfiguration: serverConfig,
	}

	_, err := service.PutBucketEncryption(input)

	return err
}

func (o *ObjectScanner) Upload(ctx context.Context, filename string, file io.ReadCloser) error {
	const op = helper.Op("repo.ObjectScanner.Upload")

	output, err := o.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(filename),
		Body:   file,
	})

	if err != nil {
		return helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	if output.UploadID == "" {
		return helper.E(op, helper.KindUnexpected, NoObjectID, CantProcessRequest)
	}

	return nil
}

func (o *ObjectScanner) Get(ctx context.Context, filename string, file io.WriterAt) error {
	const op = helper.Op("repo.ObjectScanner.Get")

	objectInput := &s3.GetObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(filename),
	}

	_, err := o.downloader.DownloadWithContext(ctx, file, objectInput)
	if err != nil {
		return helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	return nil
}
