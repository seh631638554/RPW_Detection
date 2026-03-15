package classification

import (
	httpserver "RPW_Detection/Http"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type FetchResult struct {
	LocalPath string
	Size      int64
	SHA256    string
}

type FeatureFetcher interface {
	Fetch(bucket, key string) (*FetchResult, error)
}

type MinIOFetcher struct {
	client  *s3.S3
	workDir string
}

func NewMinIOFetcher(cfg *httpserver.ObjectStorageConfig, workDir string) (*MinIOFetcher, error) {
	if cfg == nil {
		return nil, fmt.Errorf("storage config is nil")
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("create classification workdir: %w", err)
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String(cfg.Region),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(!cfg.UseSSL),
	})
	if err != nil {
		return nil, fmt.Errorf("create minio session: %w", err)
	}

	return &MinIOFetcher{
		client:  s3.New(sess),
		workDir: workDir,
	}, nil
}

func (f *MinIOFetcher) Fetch(bucket, key string) (*FetchResult, error) {
	out, err := f.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get feature object %s/%s: %w", bucket, key, err)
	}
	defer out.Body.Close()

	localPath := filepath.Join(f.workDir, filepath.Base(key))
	file, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("create local feature file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(file, hash), out.Body)
	if err != nil {
		return nil, fmt.Errorf("write feature object to local file: %w", err)
	}

	return &FetchResult{
		LocalPath: localPath,
		Size:      size,
		SHA256:    hex.EncodeToString(hash.Sum(nil)),
	}, nil
}
