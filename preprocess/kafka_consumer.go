package preprocess

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(cfg *Config) (*KafkaConsumer, error) {
	if cfg == nil {
		return nil, errors.New("preprocess config不能为空")
	}
	if len(cfg.KafkaBrokers) == 0 {
		return nil, errors.New("kafka brokers不能为空")
	}
	if cfg.ConsumerTopic == "" {
		return nil, errors.New("消费topic不能为空")
	}
	if cfg.ConsumerGroup == "" {
		return nil, errors.New("消费组不能为空")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.KafkaBrokers,
		GroupID:        cfg.ConsumerGroup,
		Topic:          cfg.ConsumerTopic,
		MinBytes:       1,
		MaxBytes:       10 * 1024 * 1024,
		MaxWait:        cfg.PollInterval,
		CommitInterval: 0,
	})

	return &KafkaConsumer{reader: reader}, nil
}

func (c *KafkaConsumer) Consume(ctx context.Context, handler func(context.Context, AudioUploadCompletedEvent) error) error {
	defer c.reader.Close()

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var event AudioUploadCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("skip invalid upload event: %v", err)
			if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
				return commitErr
			}
			continue
		}

		if err := handler(ctx, event); err != nil {
			log.Printf("preprocess event handling failed, will retry: job_id=%s err=%v", event.JobID, err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}
