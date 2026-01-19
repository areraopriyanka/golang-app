package utils

import (
	"mime/multipart"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"time"

	"braces.dev/errtrace"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var S3Client *s3.S3

func AwsConnection() {
	// Connection for aws
	var awsConfig *aws.Config
	if config.Config.Aws.AccessKeyId != "" {
		awsConfig = &aws.Config{
			Region:      aws.String(config.Config.Aws.Region),
			Credentials: credentials.NewStaticCredentials(config.Config.Aws.AccessKeyId, config.Config.Aws.SecreteKeyId, ""),
		}
	} else {
		awsConfig = &aws.Config{
			Region: aws.String(config.Config.Aws.Region),
		}
	}
	awsSession := session.Must(session.NewSession(awsConfig))
	S3Client = s3.New(awsSession)
}

func UploadFileToS3(file *multipart.FileHeader, Filename string) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	// Upload the file to S3
	_, err = S3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(config.Config.Aws.BucketName),
		Key:    aws.String(Filename),
		Body:   src,
	})
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	// Construct url of uploaded file
	fileUrl := constant.AWS_PROTOCOL + config.Config.Aws.BucketName + constant.AWS_S3_DOMAIN + Filename
	return fileUrl, nil
}

func GeneratePreSignedUrl(s3Url string) (string, error) {
	// Generate the pre-signed URL for the S3 object
	req, _ := S3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(config.Config.Aws.BucketName),
		Key:    aws.String(s3Url),
	})
	url, err := req.Presign(config.Config.Aws.PreSignedUrlExpiration * time.Millisecond)
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	// Return the generated pre-signed URL
	return url, nil
}
