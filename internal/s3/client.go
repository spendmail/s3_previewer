package s3

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"io/ioutil"
)

type Config interface {
	GetAccessKeyId() string
	GetSecretAccessKey() string
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Client struct {
	config Config
	logger Logger
	client *s3.Client
}

// New is a s3 client constructor.
func New(config Config, logger Logger) (*Client, error) {

	cfg, err := s3config.LoadDefaultConfig(
		context.TODO(),
		s3config.WithCredentialsProvider(
			credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     config.GetAccessKeyId(),
					SecretAccessKey: config.GetSecretAccessKey(),
				},
			},
		),
	)

	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)

	return &Client{
		config: config,
		logger: logger,
		client: s3Client,
	}, nil
}

func (c *Client) Download(context context.Context, bucket, key string) ([]byte, error) {

	response, err := c.client.GetObject(context, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}

	return bytes, err
}
