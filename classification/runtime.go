package classification

import (
	"context"
	"time"
)

type Service struct {
	consumer  Consumer
	fetcher   FeatureFetcher
	predictor Predictor
	publisher Publisher
}

func NewService(consumer Consumer, fetcher FeatureFetcher, predictor Predictor, publisher Publisher) *Service {
	return &Service{
		consumer:  consumer,
		fetcher:   fetcher,
		predictor: predictor,
		publisher: publisher,
	}
}

func (s *Service) Run(ctx context.Context) error {
	return s.consumer.Consume(ctx, func(ctx context.Context, event AudioClassificationTaskEvent) error {
		fetched, err := s.fetcher.Fetch(event.FeatureBucket, event.FeatureKey)
		if err != nil {
			return err
		}

		prediction, err := s.predictor.Predict(ctx, fetched.LocalPath)
		if err != nil {
			return err
		}

		return s.publisher.PublishResult(ctx, AudioClassificationResult{
			JobID:            event.JobID,
			DeviceID:         event.DeviceID,
			SourceBucket:     event.SourceBucket,
			SourceKey:        event.SourceKey,
			FeatureBucket:    event.FeatureBucket,
			FeatureKey:       event.FeatureKey,
			LocalFeaturePath: fetched.LocalPath,
			FeatureType:      event.FeatureType,
			FeatureInputs:    event.FeatureInputs,
			FeatureShape:     event.FeatureShape,
			FeatureDType:     event.FeatureDType,
			FeatureSize:      fetched.Size,
			FeatureSHA256:    fetched.SHA256,
			ModelName:        prediction.ModelName,
			PredictionLabel:  prediction.Label,
			PredictionIndex:  prediction.Index,
			Score:            prediction.Score,
			Probabilities:    prediction.Probabilities,
			ClassifiedAt:     time.Now(),
		})
	})
}
