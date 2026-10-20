package repository

import (
	"context"

	"cloud-file-ingestion/internal/file"
)

type FileRepository interface {
	Create(ctx context.Context, metadata *file.Metadata) error
}
