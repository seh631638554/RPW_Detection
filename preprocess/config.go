package preprocess

import (
	httpserver "RPW_Detection/Http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Storage       *httpserver.ObjectStorageConfig
	WorkDir       string
	ConsumerTopic string
	ResultTopic   string
	ConsumerGroup string
	PollInterval  time.Duration
	ProcessorName string
}

func LoadConfig() *Config {
	workDir := os.Getenv("PREPROCESS_WORK_DIR")
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "rpw-audio-preprocess")
	}

	return &Config{
		Storage:       httpserver.LoadObjectStorageConfig(),
		WorkDir:       workDir,
		ConsumerTopic: getEnv("PREPROCESS_INPUT_TOPIC", "audio.upload.completed"),
		ResultTopic:   getEnv("PREPROCESS_RESULT_TOPIC", "audio.preprocess.completed"),
		ConsumerGroup: getEnv("PREPROCESS_GROUP_ID", "audio_preprocess_group"),
		PollInterval:  getDurationEnv("PREPROCESS_POLL_INTERVAL", 3*time.Second),
		ProcessorName: getEnv("PREPROCESSOR_NAME", "basic-audio-preprocessor"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
