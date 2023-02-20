package repo

import (
	"backstreetlinkv2/cmd/helper"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
)

var (
	NoObjectID = errors.New("object id is not found/empty")
)

type ObjectScanner struct {
	client     *s3.Client
	bucketName string
}

type ObjectConfig struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Bucket    string
}

func NewObjectScanner(ctx context.Context, cfg ObjectConfig) (*ObjectScanner, error) {
	obj := &ObjectScanner{}
	obj.bucketName = cfg.Bucket

	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           cfg.Endpoint,
			SigningRegion: "de",
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(creds), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		return nil, err
	}

	obj.client = s3.NewFromConfig(awsCfg)

	return obj, nil
}

func (o *ObjectScanner) Upload(ctx context.Context, filename string, fileToUpload io.ReadCloser) error {
	const op = helper.Op("repo.ObjectScanner.Upload")

	defer fileToUpload.Close()

	output, err := o.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(o.bucketName),
		Key:                  aws.String(filename),
		Body:                 fileToUpload,
		ServerSideEncryption: types.ServerSideEncryptionAes256,
	})

	if err != nil {
		return helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	if output.ETag == nil {
		return helper.E(op, helper.KindUnexpected, NoObjectID, CantProcessRequest)
	}

	return nil
}

type FileStat struct {
	ContentType        string
	ContentLength      int64
	ContentDisposition string
	Fill               string
}

func (o *ObjectScanner) Get(ctx context.Context, filename string, to io.Writer) (FileStat, error) {
	const op = helper.Op("repo.ObjectScanner.Get")

	objectInput := &s3.GetObjectInput{
		Bucket:               aws.String(o.bucketName),
		Key:                  aws.String(filename),
		SSECustomerAlgorithm: aws.String("AES256"),
	}

	object, err := o.client.GetObject(ctx, objectInput)
	if err != nil {
		return FileStat{}, helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	defer object.Body.Close()

	_, err = io.Copy(to, object.Body)
	if err != nil {
		return FileStat{}, helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	fs := FileStat{
		ContentType:        aws.ToString(object.ContentType),
		ContentLength:      object.ContentLength,
		ContentDisposition: aws.ToString(object.ContentDisposition),
	}

	return fs, nil
}
