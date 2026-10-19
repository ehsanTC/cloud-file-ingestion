package main

import (
	"cloud-file-ingestion/common"
	"cloud-file-ingestion/config"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var conf *config.Config
var cfg aws.Config

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
	cfg = aws.Config{
		Region: conf.BlobStorage.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			conf.BlobStorage.AccessKey,
			conf.BlobStorage.SecretKey,
			"",
		),
	}
	return nil
}

func registerEndPoints(r *gin.Engine) {
	registerHealth(r)
	registerFileApis(r)
}

func registerFileApis(r *gin.Engine) {
	files := r.Group(fmt.Sprintf("/files/%s", common.ApiVersion1))
	{
		files.POST("", func(c *gin.Context) {
			// limit the file upload size
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(conf.App.FileSpecs.BigSize))
			fileHeader, err := c.FormFile("file")

			if err != nil {
				var maxErr *http.MaxBytesError
				if errors.As(err, &maxErr) {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "The file is too large. Use another api to upload file"})
					return
				}

				c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
				return
			}

			ext := filepath.Ext(fileHeader.Filename)
			if err := checkAllowedExtensions(ext); err != nil {
				c.JSON(http.StatusNotAcceptable, gin.H{"error": "File type is not allowed"})
				return
			}

			if err := checkFileType(fileHeader); err != nil {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
				return
			}

			if fileId, err := writeFile(fileHeader); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusCreated, gin.H{"file_id": fileId})
			}
			return
		})
	}
}

func writeFile(header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	key := uuid.New().String()
	contentType := header.Header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = "application/octet-stream"
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(conf.BlobStorage.Endpoint)
		o.UsePathStyle = true
	})
	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(conf.BlobStorage.Bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(header.Size),
	})

	if err != nil {
		return "", err
	}

	return key, nil
}

func checkFileType(header *multipart.FileHeader) error {
	src, err := header.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return errors.New("error reading file")
	}

	contentType := http.DetectContentType(buf[:n])
	return checkAllowedExtensions(contentType)
}

func checkAllowedExtensions(ext string) error {
	for _, blocked := range conf.App.FileSpecs.BlockedExtensions {
		if strings.EqualFold(blocked, ext) {
			return errors.New("file extension is not allowed")
		}
	}

	return nil
}

func registerHealth(r *gin.Engine) gin.IRoutes {
	return r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
}
