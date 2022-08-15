package s3_bucket

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gofiber/fiber/v2/utils"
)

var (
	sess     *session.Session
	uploader *s3manager.Uploader
)

type bucket struct {
}

func init() {
	sess = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
	uploader = s3manager.NewUploader(sess)
}

func NewBucket() *bucket {
	return &bucket{}
}

func (b *bucket) Upload(file io.Reader, userId string) (string, error) {
	key := fmt.Sprintf("user:%s/%s", userId, utils.UUIDv4())

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("kofi-blog-bucket"),
		Key:    aws.String(key),
		Body:   file,
	})

	if err != nil {
		return "", err
	}

	return result.Location, nil
}
