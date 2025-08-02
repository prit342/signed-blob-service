package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"
)

// Storage errors
var (
	ErrBlobNotFound = errors.New("blob not found")
	ErrBlobExists   = errors.New("blob already exists")
)

// Storage defines the interface for blob storage operations
type Storage interface {
	// Store saves a new blob to the storage
	Store(ctx context.Context, record *blobv1.SignedBlobRecord) error
	// GetByUUID retrieves a blob by its UUID
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*blobv1.SignedBlobRecord, error)
	// Exists checks if a blob with the given UUID exists
	Exists(ctx context.Context, uuid uuid.UUID) (bool, error)
	// Delete removes a blob by its UUID (optional for future use)
	Delete(ctx context.Context, uuid uuid.UUID) error
	// Migrate helps migrate database schema using migration files in the directory
	Migrate(ctx context.Context, directory string) error
	// Ping checks if the storage is reachable
	Ping(ctx context.Context) error
}
