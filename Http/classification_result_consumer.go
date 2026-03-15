package httpserver

import (
	dao "RPW_Detection/Dao"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type classificationResultMessage struct {
	JobID           string             `json:"job_id"`
	FeatureBucket   string             `json:"feature_bucket"`
	FeatureKey      string             `json:"feature_key"`
	PredictionLabel string             `json:"prediction_label"`
	Score           float64            `json:"score"`
	Probabilities   map[string]float64 `json:"probabilities"`
}

// 拿到分类结果，从kafka？？？
func StartClassificationResultConsumer(cfg *Config, repo *dao.Repo) {
	if cfg == nil || repo == nil || len(cfg.Kafka.Brokers) == 0 {
		return
	}

	topic := getEnv("CLASSIFICATION_RESULT_TOPIC", "audio.classification.result")
	groupID := getEnv("CLASSIFICATION_RESULT_GROUP_ID", "http_classification_result_group")

	go func() {
		for {
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers: cfg.Kafka.Brokers,
				Topic:   topic,
				GroupID: groupID,
			})

			func() {
				defer reader.Close()
				for {
					msg, err := reader.ReadMessage(context.Background())
					if err != nil {
						log.Printf("classification result consumer stopped: %v", err)
						return
					}

					var result classificationResultMessage
					if err := json.Unmarshal(msg.Value, &result); err != nil {
						log.Printf("classification result parse failed: %v", err)
						continue
					}
					if result.JobID == "" {
						continue
					}

					updates := map[string]interface{}{
						"status":        "done",
						"result_label":  result.PredictionLabel,
						"result_score":  result.Score,
						"error_message": "",
					}
					if result.FeatureBucket != "" {
						updates["feature_bucket"] = result.FeatureBucket
					}
					if result.FeatureKey != "" {
						updates["feature_key"] = result.FeatureKey
					}

					if err := repo.UpdateDetectionRecordByJobID(result.JobID, updates); err != nil {
						log.Printf("classification result update failed, job_id=%s err=%v", result.JobID, err)
						continue
					}
					log.Printf("classification result synced, job_id=%s label=%s score=%.6f", result.JobID, result.PredictionLabel, result.Score)
				}
			}()

			time.Sleep(2 * time.Second)
		}
	}()
}
