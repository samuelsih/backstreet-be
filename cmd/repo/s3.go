package repo

import (
	"backstreetlinkv2/cmd/model"
	"context"
	"errors"
	"math/rand"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
)


func init() {
	rand.Seed(time.Now().Unix())
}

func HandleS3(ctx context.Context, client *mongo.Client, req model.ShortenFileRequest, file *multipart.Part) (model.ShortenResponse, error) {
	var response model.ShortenResponse

	_, err := Find(ctx, client, req.Alias)
	if err == nil {
		return response, errors.New("oops, this alias has been used before, choose another one")
	}

	sess, err := connectAWS()
	if err != nil {
		return response, err
	}

	uploader := s3manager.NewUploader(sess)
	filename := file.FileName()
	filenameSource := generateFilename(file.FileName())

	params := &s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKETNAME")),
		Key:    aws.String(filenameSource),
		Body:   file,
	}

	_, err = uploader.UploadWithContext(ctx, params)

	if err != nil {
		return response, err
	}

	req.Filename = filename
	req.FilenameSource = filenameSource

	if err := InsertLink(ctx, client, req); err != nil {
		return response, err
	}

	response.Alias = req.Alias
	response.Filename = req.Filename
	response.Type = req.Type

	return response, nil
}

func HandleFindFileRequest(ctx context.Context, res model.ShortenResponse) (*s3.GetObjectOutput, error) {
	sess, err := connectAWS()
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	objectInput := &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKETNAME")),
		Key:    aws.String(res.FilenameSource),
	}

	
	fileObj, err := svc.GetObjectWithContext(ctx, objectInput)
	if err != nil {
		return nil, err
	}

	return fileObj, nil
}	

// func newSignedGetURL(sess *session.Session, objectKey string, ttl time.Duration) (string, error) {
// 	svc := s3.New(sess)

// 	objectInput := &s3.GetObjectInput{
// 		Bucket: aws.String(os.Getenv("AWS_BUCKETNAME")),
// 		Key:    aws.String(objectKey),
// 	}

// 	obj, _ := svc.GetObjectRequest(objectInput)
// 	if obj == nil {
// 		return "", errors.New("not found")
// 	}

// 	url, err := obj.Presign(ttl)
// 	if err != nil {
// 		return "", err
// 	}

// 	return url, nil
// }

func connectAWS() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"), ""),
		Endpoint:         aws.String(os.Getenv("AWS_ENDPOINT")),
		Region:           aws.String(os.Getenv("AWS_REGION")),
		S3ForcePathStyle: aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}

	return sess, nil
}

func generateFilename(str string) string {
	s := strings.Split(str, ".")

	if len(s) == 1 {
		return randomString()
	}

	s[0] = randomString()

	return strings.Join(s, ".")
}

func randomString() string {
	charset := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP1234567890"

	var output strings.Builder

	for i := 0; i < 15; i++ {
		random := rand.Intn(len(charset))
		randomChar := charset[random]
		output.WriteString(string(randomChar))
	}

	return output.String()
}