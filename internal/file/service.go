package file

import (
	"cloud-file-ingestion/config"
	"context"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileService struct {
	conf   *config.Config
	client *s3.Client
}

func NewFileService(conf *config.Config, client *s3.Client) *FileService {
	return &FileService{conf: conf, client: client}
}

func (h *FileService) Upload(c *gin.Context) {
	// limit the file upload size
	c.Request.Body = http.MaxBytesReader(c.Writer,
		c.Request.Body,
		int64(h.conf.App.FileSpecs.BigSize))

	fileHeader, err := c.FormFile("file")
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			c.JSON(
				http.StatusRequestEntityTooLarge,
				gin.H{"error": "The file is too large. Use another api to upload file"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	if h.isBlocked(filepath.Ext(fileHeader.Filename)) {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "File type is not allowed"})
		return
	}

	fileID, err := h.writeFile(c.Request.Context(), fileHeader)
	if err != nil {
		log.Printf("upload file: %v", err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Could not store the file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"file_id": fileID})
}

func (h *FileService) writeFile(ctx context.Context, header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	key := uuid.New().String()
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = h.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(h.conf.BlobStorage.Bucket),
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

func (h *FileService) isBlocked(value string) bool {
	for _, blocked := range h.conf.App.FileSpecs.BlockedExtensions {
		if strings.EqualFold(blocked, value) {
			return true
		}
	}
	return false
}
