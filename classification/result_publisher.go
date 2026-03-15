package classification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Publisher interface {
	PublishResult(ctx context.Context, result AudioClassificationResult) error
}

type KafkaResultPublisher struct {
	writer *kafka.Writer
}

func NewKafkaResultPublisher(cfg *Config) (*KafkaResultPublisher, error) {
	if cfg == nil {
		return nil, fmt.Errorf("classification config is nil")
	}
	if len(cfg.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("kafka brokers不能为空")
	}
	if cfg.ResultTopic == "" {
		return nil, fmt.Errorf("classification result topic不能为空")
	}

	return &KafkaResultPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.KafkaBrokers...),
			Topic:        cfg.ResultTopic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
		},
	}, nil
}

func (p *KafkaResultPublisher) PublishResult(ctx context.Context, result AudioClassificationResult) error {
	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal classification result: %w", err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(result.JobID),
		Value: body,
		Time:  result.ClassifiedAt,
	}); err != nil {
		return fmt.Errorf("publish classification result: %w", err)
	}

	log.Printf("audio classification result: %s", body)
	return nil
}
