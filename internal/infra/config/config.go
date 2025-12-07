package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Log        LogConfig        `yaml:"log"`
	HTTPClient HTTPClientConfig `yaml:"http_client"`
	Limiter    LimiterConfig    `yaml:"limiter"`
	Gemini     GeminiConfig     `yaml:"gemini"`
	ImageGen   ImageGenConfig   `yaml:"image_gen"`
	Storage    StorageConfig    `yaml:"storage"`
}

type ServerConfig struct {
	Addr                string `yaml:"addr"`
	ReadTimeoutSeconds  int    `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `yaml:"write_timeout_seconds"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type HTTPClientConfig struct {
	TimeoutSeconds int `yaml:"timeout_seconds"`
	MaxRetries     int `yaml:"max_retries"`
}

type LimiterConfig struct {
	MaxConcurrent int     `yaml:"max_concurrent"`
	RatePerSecond float64 `yaml:"rate_per_second"`
}

type GeminiConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type ImageGenConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type StorageConfig struct {
	Type     string `yaml:"type"`
	BasePath string `yaml:"base_path"`
	BaseURL  string `yaml:"base_url"`
}

func Load() (*Config, error) {
	cfg := defaultConfig()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return applyEnvOverrides(cfg), nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return applyEnvOverrides(cfg), nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:                ":8080",
			ReadTimeoutSeconds:  30,
			WriteTimeoutSeconds: 120,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		HTTPClient: HTTPClientConfig{
			TimeoutSeconds: 60,
			MaxRetries:     2,
		},
		Limiter: LimiterConfig{
			MaxConcurrent: 10,
			RatePerSecond: 5,
		},
		Gemini: GeminiConfig{
			Model: "gemini-3-pro-image-preview",
		},
		ImageGen: ImageGenConfig{
			Model: "gemini-3-pro-image-preview",
		},
		Storage: StorageConfig{
			Type:     "local",
			BasePath: "./output",
			BaseURL:  "/files",
		},
	}
}

func applyEnvOverrides(cfg *Config) *Config {
	if v := os.Getenv("SERVER_ADDR"); v != "" {
		cfg.Server.Addr = v
	}
	if v := os.Getenv("GEMINI_API_KEY"); v != "" {
		cfg.Gemini.APIKey = v
	}
	if v := os.Getenv("GEMINI_MODEL"); v != "" {
		cfg.Gemini.Model = v
	}
	if v := os.Getenv("IMAGEGEN_API_KEY"); v != "" {
		cfg.ImageGen.APIKey = v
	}
	if v := os.Getenv("IMAGEGEN_MODEL"); v != "" {
		cfg.ImageGen.Model = v
	}
	if v := os.Getenv("STORAGE_TYPE"); v != "" {
		cfg.Storage.Type = v
	}
	if v := os.Getenv("STORAGE_BASE_PATH"); v != "" {
		cfg.Storage.BasePath = v
	}
	if v := os.Getenv("STORAGE_BASE_URL"); v != "" {
		cfg.Storage.BaseURL = v
	}
	return cfg
}
