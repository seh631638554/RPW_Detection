package preprocess

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/segmentio/kafka-go"
)

type MinIOKafkaPublisher struct {
	bucket string
	client *s3.S3
	writer *kafka.Writer
}

func NewMinIOKafkaPublisher(cfg *Config) (*MinIOKafkaPublisher, error) {
	if cfg == nil {
		return nil, fmt.Errorf("preprocess config is nil")
	}
	if cfg.Storage == nil {
		return nil, fmt.Errorf("storage config is nil")
	}
	if len(cfg.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("kafka brokers不能为空")
	}
	if strings.TrimSpace(cfg.ResultTopic) == "" {
		return nil, fmt.Errorf("classification topic不能为空")
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Endpoint:         aws.String(cfg.Storage.Endpoint),
		Region:           aws.String(cfg.Storage.Region),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(!cfg.Storage.UseSSL),
	})
	if err != nil {
		return nil, fmt.Errorf("create minio session: %w", err)
	}

	return &MinIOKafkaPublisher{
		bucket: cfg.Storage.Bucket,
		client: s3.New(sess),
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.KafkaBrokers...),
			Topic:        cfg.ResultTopic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
		},
	}, nil
}

func (p *MinIOKafkaPublisher) PublishResult(ctx context.Context, result AudioPreprocessResult) error {
	featureKey, err := generateFeatureObjectKey(result.Key, result.FeatureType, result.FeaturePath)
	if err != nil {
		return err
	}
	if err := p.uploadFeatureFile(ctx, featureKey, result.FeaturePath); err != nil {
		return err
	}

	result.FeatureBucket = p.bucket
	result.FeatureKey = featureKey

	event := AudioClassificationTaskEvent{
		JobID:         result.JobID,
		DeviceID:      result.DeviceID,
		SourceBucket:  result.Bucket,
		SourceKey:     result.Key,
		FeatureBucket: result.FeatureBucket,
		FeatureKey:    result.FeatureKey,
		FeatureType:   result.FeatureType,
		FeatureInputs: result.FeatureInputs,
		FeatureShape:  result.FeatureShape,
		FeatureDType:  result.FeatureDType,
		SampleRate:    result.SampleRate,
		NMels:         result.NMels,
		ProcessorName: result.ProcessorName,
		CreatedAt:     time.Now(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal classification event: %w", err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(result.JobID),
		Value: body,
		Time:  event.CreatedAt,
	}); err != nil {
		return fmt.Errorf("publish classification event: %w", err)
	}

	logBody, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal preprocess result: %w", err)
	}
	log.Printf("audio preprocess result: %s", logBody)
	return nil
}

func (p *MinIOKafkaPublisher) uploadFeatureFile(ctx context.Context, key, localPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open feature file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat feature file: %w", err)
	}

	_, err = p.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(stat.Size()),
		ContentType:   aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("upload feature file to minio: %w", err)
	}
	return nil
}

func generateFeatureObjectKey(sourceKey, featureType, localPath string) (string, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return "", fmt.Errorf("source key is empty")
	}

	sourceExt := path.Ext(sourceKey)
	base := strings.TrimSuffix(sourceKey, sourceExt)

	outputExt := filepath.Ext(localPath)
	if outputExt == "" {
		outputExt = ".pt"
	}

	return path.Clean("features/" + base + "." + featureType + outputExt), nil
}
