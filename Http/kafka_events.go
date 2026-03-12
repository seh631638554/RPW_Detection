package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type AudioUploadCompletedEvent struct {
	JobID       string    `json:"job_id"`
	DeviceID    string    `json:"device_id"`
	Bucket      string    `json:"bucket"`
	Key         string    `json:"key"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type UploadEventPublisher interface {
	PublishAudioUploadCompleted(ctx context.Context, event AudioUploadCompletedEvent) error
}

type KafkaUploadEventPublisher struct {
	writer *kafka.Writer
}

var (
	uploadEventPublisher   UploadEventPublisher
	uploadEventPublisherMu sync.Mutex
)

func NewKafkaUploadEventPublisher(cfg KafkaConfig) (*KafkaUploadEventPublisher, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers不能为空")
	}
	if strings.TrimSpace(cfg.Topic) == "" {
		return nil, errors.New("kafka topic不能为空")
	}

	return &KafkaUploadEventPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.Topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}, nil
}

func (p *KafkaUploadEventPublisher) PublishAudioUploadCompleted(ctx context.Context, event AudioUploadCompletedEvent) error {
	if p == nil || p.writer == nil {
		return errors.New("kafka writer未初始化")
	}

	if event.UploadedAt.IsZero() {
		event.UploadedAt = time.Now()
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.JobID),
		Value: body,
		Time:  event.UploadedAt,
	})
}

func ensureUploadEventPublisher() error {
	uploadEventPublisherMu.Lock()
	defer uploadEventPublisherMu.Unlock()

	if uploadEventPublisher != nil {
		return nil
	}

	publisher, err := NewKafkaUploadEventPublisher(LoadConfig().Kafka)
	if err != nil {
		return err
	}
	uploadEventPublisher = publisher
	return nil
}

func extractDeviceIDFromObjectKey(key string) string {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if key == "" {
		return ""
	}
	parts := strings.SplitN(key, "/", 2)
	return parts[0]
}
