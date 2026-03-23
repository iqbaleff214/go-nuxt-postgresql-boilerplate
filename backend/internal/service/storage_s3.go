package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3StorageService stores files on any S3-compatible backend (AWS, R2, MinIO).
type S3StorageService struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
	publicURL string // optional CDN base; empty = use pre-signed URLs
}

func NewS3StorageService(ctx context.Context, endpoint, bucket, accessKey, secretKey, region, publicURL string) (*S3StorageService, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // required for MinIO
		}
	})

	return &S3StorageService{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    bucket,
		publicURL: publicURL,
	}, nil
}

func (s *S3StorageService) Upload(ctx context.Context, file io.Reader, path string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   file,
		ACL:    types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("s3 put object: %w", err)
	}
	if s.publicURL != "" {
		return s.publicURL + "/" + path, nil
	}
	// Return a long-lived pre-signed URL as fallback
	return s.GetSignedURL(ctx, path, 7*24*time.Hour)
}

func (s *S3StorageService) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (s *S3StorageService) GetSignedURL(ctx context.Context, path string, expiresIn time.Duration) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}
	return req.URL, nil
}

// NewStorageService returns the correct backend based on config.
func NewStorageService(ctx context.Context, storageBackend, storagePath, s3Endpoint, s3Bucket, s3AccessKey, s3SecretKey, s3Region, s3PublicURL string) (StorageService, error) {
	if storageBackend == "s3" {
		return NewS3StorageService(ctx, s3Endpoint, s3Bucket, s3AccessKey, s3SecretKey, s3Region, s3PublicURL)
	}
	return NewLocalStorageService(storagePath), nil
}
