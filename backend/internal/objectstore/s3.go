package objectstore

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config configures an S3-compatible object store.
type S3Config struct {
	Region          string
	Bucket          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	ForcePathStyle  bool
}

// S3Client stores objects and issues presigned uploads to S3-compatible storage.
type S3Client struct {
	bucket    string
	client    *s3.Client
	presigner *s3.PresignClient
}

func NewS3Client(ctx context.Context, cfg S3Config) (*S3Client, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("object store bucket is required")
	}
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if cfg.AccessKeyID != "" || cfg.SecretAccessKey != "" || cfg.SessionToken != "" {
		provider := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken)
		opts = append(opts, config.WithCredentialsProvider(provider))
	}

	if cfg.Endpoint != "" {
		opts = append(opts, config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, _ ...any) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:               cfg.Endpoint,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			},
		)))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle
		// Railway Buckets (S3-compatible) does not always return AWS checksum
		// response headers for GetObject. Avoid repeated SDK warnings and only
		// validate when explicitly required by a request.
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		o.DisableLogOutputChecksumValidationSkipped = true
	})

	return &S3Client{
		bucket:    cfg.Bucket,
		client:    client,
		presigner: s3.NewPresignClient(client),
	}, nil
}

func (c *S3Client) URIForKey(key string) string {
	return fmt.Sprintf("s3://%s/%s", c.bucket, strings.TrimPrefix(key, "/"))
}

func (c *S3Client) Put(ctx context.Context, key string, r io.Reader) (string, int64, error) {
	cleanKey := strings.TrimPrefix(key, "/")
	input := &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    &cleanKey,
		Body:   r,
	}

	_, err := c.client.PutObject(ctx, input)
	if err != nil {
		return "", 0, err
	}

	return c.URIForKey(cleanKey), 0, nil
}

func (c *S3Client) PresignPut(ctx context.Context, key, contentType string) (string, map[string]string, string, error) {
	cleanKey := strings.TrimPrefix(key, "/")
	input := &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    &cleanKey,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	result, err := c.presigner.PresignPutObject(ctx, input)
	if err != nil {
		return "", nil, "", err
	}

	headers := map[string]string{}
	for key, values := range result.SignedHeader {
		headers[key] = strings.Join(values, ",")
	}

	return result.URL, headers, c.URIForKey(cleanKey), nil
}

func (c *S3Client) Get(ctx context.Context, uri string) (io.ReadCloser, int64, error) {
	bucket, key, err := parseS3URI(uri)
	if err != nil {
		return nil, 0, err
	}
	if bucket != c.bucket {
		return nil, 0, fmt.Errorf("s3 uri bucket mismatch: %s", bucket)
	}

	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	output, err := c.client.GetObject(ctx, input)
	if err != nil {
		return nil, 0, err
	}

	size := int64(0)
	if output.ContentLength != nil {
		size = *output.ContentLength
	}

	return output.Body, size, nil
}

func parseS3URI(uri string) (string, string, error) {
	if !strings.HasPrefix(uri, "s3://") {
		return "", "", fmt.Errorf("invalid s3 uri: %s", uri)
	}
	trimmed := strings.TrimPrefix(uri, "s3://")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid s3 uri: %s", uri)
	}
	return parts[0], parts[1], nil
}

var _ Client = (*S3Client)(nil)
