package preprocess

import (
	"context"
	"path/filepath"
	"time"
)

type Processor interface {
	Process(ctx context.Context, event AudioUploadCompletedEvent) (*AudioPreprocessResult, error)
}

type BasicProcessor struct {
	fetcher   ObjectFetcher
	extractor FeatureExtractor
	name      string
}

func NewBasicProcessor(fetcher ObjectFetcher, extractor FeatureExtractor, name string) *BasicProcessor {
	return &BasicProcessor{
		fetcher:   fetcher,
		extractor: extractor,
		name:      name,
	}
}

func (p *BasicProcessor) Process(ctx context.Context, event AudioUploadCompletedEvent) (*AudioPreprocessResult, error) {
	_ = ctx

	fetched, err := p.fetcher.Fetch(event.Bucket, event.Key)
	if err != nil {
		return nil, err
	}
	features, err := p.extractor.Extract(ctx, fetched.LocalPath)
	if err != nil {
		return nil, err
	}

	return &AudioPreprocessResult{
		JobID:         event.JobID,
		DeviceID:      event.DeviceID,
		Bucket:        event.Bucket,
		Key:           event.Key,
		LocalPath:     fetched.LocalPath,
		ObjectSize:    fetched.Size,
		SHA256:        fetched.SHA256,
		ContentType:   event.ContentType,
		FileExt:       filepath.Ext(event.Key),
		FeaturePath:   features.OutputPath,
		FeatureType:   features.FeatureType,
		FeatureInputs: features.Inputs,
		FeatureShape:  features.Shape,
		FeatureDType:  features.DType,
		SampleRate:    features.SampleRate,
		NMels:         features.NMels,
		ProcessedAt:   time.Now(),
		ProcessorName: p.name,
	}, nil
}
