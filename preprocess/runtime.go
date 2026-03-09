package preprocess

import (
	"context"
	"encoding/json"
	"log"
)

type Consumer interface {
	Consume(ctx context.Context, handler func(context.Context, AudioUploadCompletedEvent) error) error
}

type Publisher interface {
	PublishResult(ctx context.Context, result AudioPreprocessResult) error
}

type Service struct {
	consumer  Consumer
	processor Processor
	publisher Publisher
}

func NewService(consumer Consumer, processor Processor, publisher Publisher) *Service {
	return &Service{
		consumer:  consumer,
		processor: processor,
		publisher: publisher,
	}
}

func (s *Service) Run(ctx context.Context) error {
	return s.consumer.Consume(ctx, func(ctx context.Context, event AudioUploadCompletedEvent) error {
		result, err := s.processor.Process(ctx, event)
		if err != nil {
			return err
		}
		return s.publisher.PublishResult(ctx, *result)
	})
}

type NoopConsumer struct{}

func (NoopConsumer) Consume(ctx context.Context, handler func(context.Context, AudioUploadCompletedEvent) error) error {
	_ = handler
	log.Println("preprocess consumer is in noop mode; kafka adapter not connected yet")
	<-ctx.Done()
	return ctx.Err()
}

type LogPublisher struct{}

func (LogPublisher) PublishResult(ctx context.Context, result AudioPreprocessResult) error {
	_ = ctx
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}
	log.Printf("audio preprocess result: %s", body)
	return nil
}

