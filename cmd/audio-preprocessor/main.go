package main

import (
	"RPW_Detection/preprocess"
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
	cfg := preprocess.LoadConfig()

	once := flag.Bool("once", false, "process a single object and exit")
	jobID := flag.String("job-id", "", "job id for --once mode")
	deviceID := flag.String("device-id", "", "device id for --once mode")
	bucket := flag.String("bucket", cfg.Storage.Bucket, "object bucket")
	key := flag.String("key", "", "object key for --once mode")
	size := flag.Int64("size", 0, "object size for --once mode")
	contentType := flag.String("content-type", "application/octet-stream", "content type for --once mode")
	flag.Parse()

	fetcher, err := preprocess.NewMinIOFetcher(cfg.Storage, cfg.WorkDir)
	if err != nil {
		log.Fatalf("init minio fetcher failed: %v", err)
	}
	processor := preprocess.NewBasicProcessor(fetcher, cfg.ProcessorName)

	if *once {
		if *jobID == "" || *bucket == "" || *key == "" {
			log.Fatal("--once requires --job-id, --bucket and --key")
		}
		runOnce(processor, preprocess.AudioUploadCompletedEvent{
			JobID:       *jobID,
			DeviceID:    *deviceID,
			Bucket:      *bucket,
			Key:         *key,
			Size:        *size,
			ContentType: *contentType,
			UploadedAt:  time.Now(),
		})
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	service := preprocess.NewService(preprocess.NoopConsumer{}, processor, preprocess.LogPublisher{})
	if err := service.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("preprocess service stopped with error: %v", err)
	}
}

func runOnce(processor preprocess.Processor, event preprocess.AudioUploadCompletedEvent) {
	result, err := processor.Process(context.Background(), event)
	if err != nil {
		log.Fatalf("process object failed: %v", err)
	}

	body, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("marshal result failed: %v", err)
	}
	log.Println(string(body))
}
