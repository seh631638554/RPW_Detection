package main

import (
	"RPW_Detection/classification"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := classification.LoadConfig()

	once := flag.Bool("once", false, "classify a single feature object and exit")
	jobID := flag.String("job-id", "", "job id for --once mode")
	deviceID := flag.String("device-id", "", "device id for --once mode")
	sourceBucket := flag.String("source-bucket", cfg.Storage.Bucket, "source audio bucket")
	sourceKey := flag.String("source-key", "", "source audio key")
	featureBucket := flag.String("feature-bucket", cfg.Storage.Bucket, "feature bucket")
	featureKey := flag.String("feature-key", "", "feature key for --once mode")
	flag.Parse()

	fetcher, err := classification.NewMinIOFetcher(cfg.Storage, cfg.WorkDir)
	if err != nil {
		log.Fatalf("init minio fetcher failed: %v", err)
	}
	predictor := classification.NewPythonPredictor(cfg)

	if *once {
		if *jobID == "" || *featureBucket == "" || *featureKey == "" {
			log.Fatal("--once requires --job-id, --feature-bucket and --feature-key")
		}
		runOnce(fetcher, predictor, classification.AudioClassificationTaskEvent{
			JobID:         *jobID,
			DeviceID:      *deviceID,
			SourceBucket:  *sourceBucket,
			SourceKey:     *sourceKey,
			FeatureBucket: *featureBucket,
			FeatureKey:    *featureKey,
			CreatedAt:     time.Now(),
		})
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	consumer, err := classification.NewKafkaConsumer(cfg)
	if err != nil {
		log.Fatalf("init kafka consumer failed: %v", err)
	}
	publisher, err := classification.NewKafkaResultPublisher(cfg)
	if err != nil {
		log.Fatalf("init result publisher failed: %v", err)
	}

	service := classification.NewService(consumer, fetcher, predictor, publisher)
	if err := service.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("classification service stopped with error: %v", err)
	}
}

func runOnce(fetcher classification.FeatureFetcher, predictor classification.Predictor, event classification.AudioClassificationTaskEvent) {
	fetched, err := fetcher.Fetch(event.FeatureBucket, event.FeatureKey)
	if err != nil {
		log.Fatalf("fetch feature failed: %v", err)
	}

	prediction, err := predictor.Predict(context.Background(), fetched.LocalPath)
	if err != nil {
		log.Fatalf("predict feature failed: %v", err)
	}

	result := classification.AudioClassificationResult{
		JobID:            event.JobID,
		DeviceID:         event.DeviceID,
		SourceBucket:     event.SourceBucket,
		SourceKey:        event.SourceKey,
		FeatureBucket:    event.FeatureBucket,
		FeatureKey:       event.FeatureKey,
		LocalFeaturePath: fetched.LocalPath,
		FeatureSize:      fetched.Size,
		FeatureSHA256:    fetched.SHA256,
		ModelName:        prediction.ModelName,
		PredictionLabel:  prediction.Label,
		PredictionIndex:  prediction.Index,
		Score:            prediction.Score,
		Probabilities:    prediction.Probabilities,
		ClassifiedAt:     time.Now(),
	}

	body, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("marshal result failed: %v", err)
	}
	log.Println(string(body))
}
