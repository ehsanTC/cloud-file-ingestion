package file

import (
	"time"

	"github.com/google/uuid"
)

type Metadata struct {
	ID           uuid.UUID
	OriginalName string
	Extension    string
	ObjectKey    string
	Bucket       string
	ContentType  string
	SizeBytes    int64
	CreatedAt    time.Time
}
