package main

import (
	"cloud-file-ingestion/config"
	"cloud-file-ingestion/internal/api"
	"cloud-file-ingestion/internal/file"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

var conf *config.Config

func main() {
	if err := loadConfiguration(); err != nil {
		fmt.Println(err)
		return
	}

	services := []api.Registrable{
		file.NewFileHandler(conf, newS3Client(conf)),
	}

	r := gin.Default()
	api.RegisterEndPoints(r, services)

	address := fmt.Sprintf("%s:%d", conf.App.Server.Host, conf.App.Server.Port)
	err := r.Run(address)

	if err != nil {
		fmt.Println(err)
		return
	}
}

func loadConfiguration() error {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	configPath := ""
	for directory := workingDirectory; ; directory = filepath.Dir(directory) {
		candidate := filepath.Join(directory, "config", "configurations.yml")
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			configPath = candidate
			break
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("check configuration file %q: %w", candidate, err)
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
	}
	if configPath == "" {
		return fmt.Errorf("config/configurations.yml not found from working directory %q", workingDirectory)
	}

	conf, err = config.Load(configPath)
	if err != nil {

		return err
	}

	return nil
}

func newS3Client(conf *config.Config) *s3.Client {
	cfg := aws.Config{
		Region: conf.BlobStorage.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			conf.BlobStorage.AccessKey,
			conf.BlobStorage.SecretKey,
			"",
		),
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(conf.BlobStorage.Endpoint)
		o.UsePathStyle = true
	})
}
