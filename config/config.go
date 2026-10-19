package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App         AppConfig         `mapstructure:"app"`
	BlobStorage BlobStorageConfig `mapstructure:"blob_storage"`
}

type AppConfig struct {
	Name      string       `mapstructure:"name"`
	Env       string       `mapstructure:"env"`
	Version   string       `mapstructure:"version"`
	Server    ServerConfig `mapstructure:"server"`
	Log       LogConfig    `mapstructure:"log"`
	FileSpecs FileSpecs    `mapstructure:"file_specs"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type FileSpecs struct {
	BigSize           int      `mapstructure:"big"`
	BlockedExtensions []string `mapstructure:"blocked_extensions"`
}

type BlobStorageConfig struct {
	Type      string `mapstructure:"type"`
	Endpoint  string `mapstructure:"endpoint"`
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// Load reads the config file into a typed struct.
// Env vars override file values, e.g. APP_SERVER_PORT overrides app.server.port.
func Load(fileName string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(fileName)
	if ext := strings.TrimPrefix(fileName[strings.LastIndex(fileName, "."):], "."); ext != "" {
		v.SetConfigType(ext)
	}

	// Env overrides: APP_SERVER_PORT, BLOB_STORAGE_SECRET_KEY, KAFKA_BROKERS, ...
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: read %q: %w", fileName, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	return &cfg, nil
}

// Read keeps your existing dotted-key accessor for ad-hoc lookups.
func Read(fileName, key string) (interface{}, error) {
	v := viper.New()
	v.SetConfigFile(fileName)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: read %q: %w", fileName, err)
	}
	if !v.IsSet(key) {
		return nil, fmt.Errorf("config: key %q not found", key)
	}
	return v.Get(key), nil
}
