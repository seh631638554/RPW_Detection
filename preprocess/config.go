package preprocess

import (
	httpserver "RPW_Detection/Http"
	"os"
	"path/filepath"
	"strconv"
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
	ProcessorName string
	PythonBin     string
	ScriptPath    string
	FeatureType   string
	SampleRate    int
	NMels         int
}

func LoadConfig() *Config {
	appCfg := httpserver.LoadConfig()
	workDir := os.Getenv("PREPROCESS_WORK_DIR")
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "rpw-audio-preprocess")
	}

	return &Config{
		Storage:       httpserver.LoadObjectStorageConfig(),
		KafkaBrokers:  appCfg.Kafka.Brokers,
		WorkDir:       workDir,
		ConsumerTopic: getEnv("PREPROCESS_INPUT_TOPIC", appCfg.Kafka.Topic),
		ResultTopic:   getEnv("PREPROCESS_RESULT_TOPIC", "audio.classification"),
		ConsumerGroup: getEnv("PREPROCESS_GROUP_ID", "audio_preprocess_group"),
		PollInterval:  getDurationEnv("PREPROCESS_POLL_INTERVAL", 3*time.Second),
		ProcessorName: getEnv("PREPROCESSOR_NAME", "basic-audio-preprocessor"),
		PythonBin:     getEnv("PREPROCESS_PYTHON_BIN", "python3"),
		ScriptPath:    getEnv("PREPROCESS_SCRIPT_PATH", filepath.Join("preprocess", "feature_extract.py")),
		FeatureType:   getEnv("PREPROCESS_FEATURE_TYPE", "log_mel_pcen"),
		SampleRate:    getIntEnv("PREPROCESS_SAMPLE_RATE", 16000),
		NMels:         getIntEnv("PREPROCESS_N_MELS", 64),
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

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
