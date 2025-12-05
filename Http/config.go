package httpserver

import (
	"os"
	"strconv"
	"time"
)

// 配置结构体
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Kafka    KafkaConfig
}

// 服务器配置
type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// 数据库配置
type DatabaseConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
	ConnLifetime time.Duration
}

// Redis配置
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
}

// Kafka配置
type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// 从环境变量加载配置
func LoadConfig() *Config {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("SERVER_MODE", "release"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getIntEnv("DB_PORT", 3306),
			Username:     getEnv("DB_USERNAME", "root"),
			Password:     getEnv("DB_PASSWORD", ""),
			Database:     getEnv("DB_DATABASE", "pest_detection"),
			MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 10),
			ConnLifetime: getDurationEnv("DB_CONN_LIFETIME", 1*time.Hour),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getIntEnv("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getIntEnv("REDIS_DATABASE", 0),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
		},
		JWT: JWTConfig{
			SecretKey:  getEnv("JWT_SECRET_KEY", "your-secret-key-here"),
			ExpireTime: getDurationEnv("JWT_EXPIRE_TIME", 24*time.Hour),
		},
		Kafka: KafkaConfig{
			Brokers: getStringSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
			Topic:   getEnv("KAFKA_TOPIC", "audio_detection"),
			GroupID: getEnv("KAFKA_GROUP_ID", "detection_group"),
		},
	}

	return config
}

// 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 获取整数环境变量
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 获取时间间隔环境变量
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// 获取字符串切片环境变量
func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// 简单的逗号分隔实现
		// 在生产环境中可能需要更复杂的解析逻辑
		return []string{value}
	}
	return defaultValue
}

// 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return c.Username + ":" + c.Password + "@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" + c.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// 获取Redis地址
func (c *RedisConfig) GetAddr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

// 获取Kafka brokers字符串
func (c *KafkaConfig) GetBrokersString() string {
	result := ""
	for i, broker := range c.Brokers {
		if i > 0 {
			result += ","
		}
		result += broker
	}
	return result
}
