package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Config는 애플리케이션 설정 구조체입니다
type Config struct {
	// 서버 설정
	Port string `json:"port"`

	// 데이터 저장 설정
	DataDir string `json:"data_dir"`

	// API 설정
	AlloraBaseURL     string `json:"allora_base_url"`
	APITimeoutSeconds int    `json:"api_timeout_seconds"`

	// 모니터링 설정
	MonitoringIntervalMinutes int `json:"monitoring_interval_minutes"`
	DataRetentionDays         int `json:"data_retention_days"`

	// 토픽 추론 데이터 설정
	TopicUpdateIntervalMinutes int      `json:"topic_update_interval_minutes"`
	DefaultActiveTopics        []string `json:"default_active_topics"`
}

// LoadConfig는 JSON 파일에서 설정을 로드합니다
func LoadConfig(path string) (*Config, error) {
	// 파일 읽기
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}

	// JSON 파싱
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("설정 파일 파싱 실패: %w", err)
	}

	// 기본값 설정
	if config.Port == "" {
		config.Port = "8080"
	}

	if config.DataDir == "" {
		config.DataDir = "data"
	}

	if config.AlloraBaseURL == "" {
		config.AlloraBaseURL = "https://forge.allora.network"
	}

	if config.APITimeoutSeconds <= 0 {
		config.APITimeoutSeconds = 30
	}

	if config.MonitoringIntervalMinutes <= 0 {
		config.MonitoringIntervalMinutes = 60
	}

	if config.DataRetentionDays <= 0 {
		config.DataRetentionDays = 30
	}

	if config.TopicUpdateIntervalMinutes <= 0 {
		config.TopicUpdateIntervalMinutes = 5
	}

	return &config, nil
}

// SaveConfig는 설정을 JSON 파일로 저장합니다
func SaveConfig(config *Config, path string) error {
	// JSON 마샬링
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("설정 마샬링 실패: %w", err)
	}

	// 파일 쓰기
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("설정 파일 쓰기 실패: %w", err)
	}

	return nil
}

// CreateDefaultConfig는 기본 설정 파일을 생성합니다
func CreateDefaultConfig(path string) error {
	config := &Config{
		Port:                       "8080",
		DataDir:                    "data",
		AlloraBaseURL:              "https://forge.allora.network",
		APITimeoutSeconds:          30,
		MonitoringIntervalMinutes:  60,
		DataRetentionDays:          30,
		TopicUpdateIntervalMinutes: 5,
		DefaultActiveTopics:        []string{},
	}

	return SaveConfig(config, path)
}

// NewConfig 환경 변수에서 설정을 로드하여 Config 구조체를 생성합니다
func NewConfig() *Config {
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		port = 8080
	}

	apiTimeout, err := strconv.Atoi(getEnv("API_REQUEST_TIMEOUT_SECONDS", "30"))
	if err != nil {
		apiTimeout = 30
	}

	monitoringInterval, err := strconv.Atoi(getEnv("MONITORING_INTERVAL_MINUTES", "60"))
	if err != nil {
		monitoringInterval = 60
	}

	dataRetention, err := strconv.Atoi(getEnv("DATA_RETENTION_DAYS", "90"))
	if err != nil {
		dataRetention = 90
	}

	return &Config{
		Port:                       strconv.Itoa(port),
		DataDir:                    getEnv("DATA_DIR", "data"),
		AlloraBaseURL:              getEnv("ALLORA_API_BASE_URL", "https://forge.allora.network"),
		APITimeoutSeconds:          apiTimeout,
		MonitoringIntervalMinutes:  monitoringInterval,
		DataRetentionDays:          dataRetention,
		TopicUpdateIntervalMinutes: 5,
		DefaultActiveTopics:        []string{},
	}
}

// getEnv 환경 변수를 가져오거나 기본값을 반환합니다
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
