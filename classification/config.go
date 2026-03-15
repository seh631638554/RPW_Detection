package classification

import (
	httpserver "RPW_Detection/Http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Storage       *httpserver.ObjectStorageConfig
	KafkaBrokers  []string
	WorkDir       string
	ConsumerTopic string
	ResultTopic   string
	ConsumerGroup string
	PollInterval  time.Duration
	PythonBin     string
	ScriptPath    string
	ModelPath     string
	Device        string
	Labels        []string
}

func LoadConfig() *Config {
	appCfg := httpserver.LoadConfig()

	workDir := os.Getenv("CLASSIFICATION_WORK_DIR")
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "rpw-audio-classification")
	}

	return &Config{
		Storage:       httpserver.LoadObjectStorageConfig(),
		KafkaBrokers:  appCfg.Kafka.Brokers,
		WorkDir:       workDir,
		ConsumerTopic: getEnv("CLASSIFICATION_INPUT_TOPIC", "audio.classification"),
		ResultTopic:   getEnv("CLASSIFICATION_RESULT_TOPIC", "audio.classification.result"),
		ConsumerGroup: getEnv("CLASSIFICATION_GROUP_ID", "audio_classification_group"),
		PollInterval:  getDurationEnv("CLASSIFICATION_POLL_INTERVAL", 3*time.Second),
		PythonBin:     getEnv("CLASSIFICATION_PYTHON_BIN", "python3"),
		ScriptPath:    getEnv("CLASSIFICATION_SCRIPT_PATH", filepath.Join("classification", "infer.py")),
		ModelPath:     getEnv("CLASSIFIER_MODEL_PATH", filepath.Join("classification", "best_fold7.pth")),
		Device:        getEnv("CLASSIFICATION_DEVICE", "cpu"),
		Labels:        getStringSliceEnv("CLASSIFIER_LABELS", []string{"clean", "infested"}),
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

func getStringSliceEnv(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				result = append(result, part)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return fallback
}
