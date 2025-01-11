package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Vision   VisionConfig   `mapstructure:"vision"`
	Image    ImageConfig    `mapstructure:"image"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

type ServerConfig struct {
	Port            int    `mapstructure:"port"`
	Host            string `mapstructure:"host"`
	Mode            string `mapstructure:"mode"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

type VisionConfig struct {
	MaxRetries      int `mapstructure:"max_retries"`
	BatchSize       int `mapstructure:"batch_size"`
	PoolSize        int `mapstructure:"pool_size"`
	RateLimit       int `mapstructure:"rate_limit"`
	TimeoutSeconds  int `mapstructure:"timeout_seconds"`
}

type ImageConfig struct {
	MaxSizeMB      int   `mapstructure:"max_size_mb"`
	MaxWidth       int   `mapstructure:"max_width"`
	MaxHeight      int   `mapstructure:"max_height"`
	Quality        int   `mapstructure:"quality"`
	AllowedFormats []string `mapstructure:"allowed_formats"`
}

type StorageConfig struct {
	OutputDir string `mapstructure:"output_dir"`
	TempDir   string `mapstructure:"temp_dir"`
}

// Load reads the configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	var config Config

	viper.SetConfigFile(configPath)
	viper.SetEnvPrefix("VISION")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.shutdown_timeout", 5)

	// Vision API defaults
	viper.SetDefault("vision.max_retries", 3)
	viper.SetDefault("vision.batch_size", 100)
	viper.SetDefault("vision.pool_size", 8)
	viper.SetDefault("vision.rate_limit", 1800)
	viper.SetDefault("vision.timeout_seconds", 30)

	// Image processing defaults
	viper.SetDefault("image.max_size_mb", 40)
	viper.SetDefault("image.max_width", 4096)
	viper.SetDefault("image.max_height", 4096)
	viper.SetDefault("image.quality", 85)
	viper.SetDefault("image.allowed_formats", []string{"jpeg", "jpg", "png", "gif", "bmp"})

	// Storage defaults
	viper.SetDefault("storage.output_dir", "./output")
	viper.SetDefault("storage.temp_dir", "./tmp")
}

func validateConfig(config *Config) error {
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", config.Server.Port)
	}

	if config.Vision.PoolSize < 1 {
		return fmt.Errorf("pool size must be at least 1")
	}

	if config.Vision.BatchSize < 1 {
		return fmt.Errorf("batch size must be at least 1")
	}

	if config.Vision.RateLimit < 1 {
		return fmt.Errorf("rate limit must be at least 1")
	}

	if config.Image.MaxSizeMB < 1 {
		return fmt.Errorf("max image size must be at least 1MB")
	}

	if config.Image.Quality < 1 || config.Image.Quality > 100 {
		return fmt.Errorf("image quality must be between 1 and 100")
	}

	if len(config.Image.AllowedFormats) == 0 {
		return fmt.Errorf("at least one image format must be allowed")
	}

	return nil
}