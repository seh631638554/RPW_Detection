package preprocess

import "time"

type AudioUploadCompletedEvent struct {
	JobID       string    `json:"job_id"`
	DeviceID    string    `json:"device_id"`
	Bucket      string    `json:"bucket"`
	Key         string    `json:"key"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type AudioPreprocessResult struct {
	JobID         string    `json:"job_id"`
	DeviceID      string    `json:"device_id"`
	Bucket        string    `json:"bucket"`
	Key           string    `json:"key"`
	LocalPath     string    `json:"local_path"`
	ObjectSize    int64     `json:"object_size"`
	SHA256        string    `json:"sha256"`
	ContentType   string    `json:"content_type"`
	FileExt       string    `json:"file_ext"`
	FeaturePath   string    `json:"feature_path"`
	FeatureBucket string    `json:"feature_bucket,omitempty"`
	FeatureKey    string    `json:"feature_key,omitempty"`
	FeatureType   string    `json:"feature_type"`
	FeatureInputs []string  `json:"feature_inputs,omitempty"`
	FeatureShape  []int     `json:"feature_shape"`
	FeatureDType  string    `json:"feature_dtype"`
	SampleRate    int       `json:"sample_rate"`
	NMels         int       `json:"n_mels"`
	ProcessedAt   time.Time `json:"processed_at"`
	ProcessorName string    `json:"processor_name"`
}

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
