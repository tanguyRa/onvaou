package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/tanguyRa/onvaou/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStore struct {
	client     *minio.Client
	bucketName string
	publicBase string
}

func NewMinIOStore(cfg config.Config) (*MinIOStore, error) {
	if cfg.Storage.MinIO.Endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT is required")
	}
	if cfg.Storage.MinIO.AccessKey == "" || cfg.Storage.MinIO.SecretKey == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY and MINIO_SECRET_KEY are required")
	}
	bucket := cfg.Storage.MinIO.Bucket
	if bucket == "" {
		bucket = "bucket"
	}

	endpoint := strings.TrimSpace(cfg.Storage.MinIO.Endpoint)
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		if parsed, err := url.Parse(endpoint); err == nil && parsed.Host != "" {
			endpoint = parsed.Host
		}
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.MinIO.AccessKey, cfg.Storage.MinIO.SecretKey, ""),
		Secure: cfg.Storage.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init MinIO client: %w", err)
	}

	store := &MinIOStore{
		client:     client,
		bucketName: bucket,
		publicBase: strings.TrimRight(cfg.Storage.MinIO.PublicBase, "/"),
	}

	if err := store.ensureBucket(context.Background()); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *MinIOStore) Put(ctx context.Context, key string, contentType string, data []byte) (string, error) {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, s.bucketName, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	if s.publicBase != "" {
		return fmt.Sprintf("%s/%s", s.publicBase, key), nil
	}

	return fmt.Sprintf("%s/%s", s.bucketName, key), nil
}

func (s *MinIOStore) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}
	return data, nil
}

func (s *MinIOStore) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if exists {
		return s.ensurePublicPolicy(ctx)
	}

	if err := s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}
	return s.ensurePublicPolicy(ctx)
}

func (s *MinIOStore) ensurePublicPolicy(ctx context.Context) error {
	policy := fmt.Sprintf(`{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetObject"],
      "Resource":["arn:aws:s3:::%s/*"]
    }
  ]
}`, s.bucketName)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if err := s.client.SetBucketPolicy(ctx, s.bucketName, policy); err == nil {
			return nil
		} else {
			lastErr = err
			if strings.Contains(err.Error(), "resource deadlock avoided") {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return fmt.Errorf("set bucket policy: %w", err)
		}
	}
	if lastErr != nil && strings.Contains(lastErr.Error(), "resource deadlock avoided") {
		return nil
	}
	if lastErr != nil {
		return fmt.Errorf("set bucket policy: %w", lastErr)
	}
	return nil
}
