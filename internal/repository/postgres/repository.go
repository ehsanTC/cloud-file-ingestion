package postgres

import (
	"context"
	"time"

	"cloud-file-ingestion/internal/file"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fileModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	OriginalName string    `gorm:"type:varchar(255);not null"`
	Extension    string    `gorm:"type:varchar(255);not null"`
	ObjectKey    string    `gorm:"type:varchar(512);not null;uniqueIndex"`
	Bucket       string    `gorm:"type:varchar(255);not null"`
	ContentType  string    `gorm:"type:varchar(255);not null"`
	SizeBytes    int64     `gorm:"not null;check:size_bytes >= 0"`
	CreatedAt    time.Time `gorm:"not null;index:idx_files_created_at,sort:desc"`
}

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(
	ctx context.Context,
	metadata *file.Metadata,
) error {
	model := fileModel{
		ID:           metadata.ID,
		OriginalName: metadata.OriginalName,
		Extension:    metadata.Extension,
		ObjectKey:    metadata.ObjectKey,
		Bucket:       metadata.Bucket,
		ContentType:  metadata.ContentType,
		SizeBytes:    metadata.SizeBytes,
		CreatedAt:    metadata.CreatedAt,
	}

	return r.db.WithContext(ctx).Create(&model).Error
}
