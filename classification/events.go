package classification

import "time"

type AudioClassificationTaskEvent struct {
	JobID         string    `json:"job_id"`
	DeviceID      string    `json:"device_id"`
	SourceBucket  string    `json:"source_bucket"`
	SourceKey     string    `json:"source_key"`
	FeatureBucket string    `json:"feature_bucket"`
	FeatureKey    string    `json:"feature_key"`
	FeatureType   string    `json:"feature_type"`
	FeatureInputs []string  `json:"feature_inputs,omitempty"`
	FeatureShape  []int     `json:"feature_shape"`
	FeatureDType  string    `json:"feature_dtype"`
	SampleRate    int       `json:"sample_rate"`
	NMels         int       `json:"n_mels"`
	ProcessorName string    `json:"processor_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type AudioClassificationResult struct {
	JobID            string             `json:"job_id"`
	DeviceID         string             `json:"device_id"`
	SourceBucket     string             `json:"source_bucket"`
	SourceKey        string             `json:"source_key"`
	FeatureBucket    string             `json:"feature_bucket"`
	FeatureKey       string             `json:"feature_key"`
	LocalFeaturePath string             `json:"local_feature_path"`
	FeatureType      string             `json:"feature_type"`
	FeatureInputs    []string           `json:"feature_inputs,omitempty"`
	FeatureShape     []int              `json:"feature_shape"`
	FeatureDType     string             `json:"feature_dtype"`
	FeatureSize      int64              `json:"feature_size"`
	FeatureSHA256    string             `json:"feature_sha256"`
	ModelName        string             `json:"model_name"`
	PredictionLabel  string             `json:"prediction_label"`
	PredictionIndex  int                `json:"prediction_index"`
	Score            float64            `json:"score"`
	Probabilities    map[string]float64 `json:"probabilities"`
	ClassifiedAt     time.Time          `json:"classified_at"`
}
