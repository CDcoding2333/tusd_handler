package mclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/minio/minio-go"
)

//Config ...
type Config struct {
	MinioEndpoint  string
	MinioRegion    string
	MinioAppID     string
	MinioAppSecret string
	MinioToken     string
}

//NewS3Client ...
func NewS3Client(conf *Config) *s3.S3 {
	// Configure to use Minio Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(conf.MinioAppID, conf.MinioAppSecret, conf.MinioToken),
		Endpoint:         aws.String(conf.MinioEndpoint),
		Region:           aws.String(conf.MinioRegion),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)

	s3Client := s3.New(newSession)
	return s3Client
}

//NewMinioClient ...
func NewMinioClient(conf *Config) *minio.Client {
	// Initialize minio client object.
	minioClient, err := minio.New(conf.MinioEndpoint, conf.MinioAppID, conf.MinioAppSecret, false)
	if err != nil {
		return nil
	}

	return minioClient
}
