package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // postgres driver
	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"
)

const (
	selectTimeQuery = `SELECT NOW()`
)

// PostgresStorage implements the Storage interface for PostgreSQL
type PostgresStorage struct {
	db  *sql.DB
	log *slog.Logger
}

// NewPostgresStorage creates a new PostgreSQL storage implementation
func NewPostgresStorage(
	dsn string, // Data Source Name for PostgreSQL connection
	log *slog.Logger, // Logger for logging
	retryInterval time.Duration, // retryInterval for pinging the database
	maxReadyDuration time.Duration, // Maximum duration to wait for the database to be ready
) (*PostgresStorage, error) {
	// validate the DSN and logger
	if dsn == "" {
		return nil, errors.New("DataSourceName (DNS) parameter cannot be empty")
	}
	if log == nil {
		return nil, errors.New("log parameter cannot be nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), maxReadyDuration)
	defer cancel()
	// we expect the database to be ready within maxReadyDuration
	db, err := pingWithRetry(ctx, dsn, retryInterval)
	if err != nil {
		log.Error("database was not ready", "error", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{db: db, log: log}, nil
}

// Store saves a new blob to the database
func (s *PostgresStorage) Store(ctx context.Context, record *blobv1.SignedBlobRecord) error {
	query := `
		INSERT INTO signed_blobs (uuid, blob, hash, timestamp, signature)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.ExecContext(ctx, query,
		record.Payload.Uuid,
		record.Payload.Blob,
		record.Payload.Hash,
		record.Payload.Timestamp,
		record.Signature, // signature is a byte slice
	)

	if err != nil {
		s.log.Error("failed to store blob", "error", err)
	}

	return err
}

// GetByUUID retrieves a blob by its UUID
func (s *PostgresStorage) GetByUUID(ctx context.Context, uuid uuid.UUID) (*blobv1.SignedBlobRecord, error) {
	query := `
		SELECT uuid, blob, hash, timestamp, signature
		FROM signed_blobs
		WHERE uuid = $1
	`

	record := &blobv1.SignedBlobRecord{
		Payload: &blobv1.BlobRecord{},
	}
	err := s.db.QueryRowContext(ctx, query, uuid).Scan(
		&record.Payload.Uuid,
		&record.Payload.Blob,
		&record.Payload.Hash,
		&record.Payload.Timestamp,
		&record.Signature,
	)

	if err != nil {
		s.log.Error("failed to retrieve blob", "error", err, "uuid", uuid)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBlobNotFound
		}
		return nil, err
	}

	return record, nil
}

// Exists checks if a blob with the given UUID exists
func (s *PostgresStorage) Exists(ctx context.Context, uuid uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM signed_blobs WHERE uuid = $1)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, uuid).Scan(&exists)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Delete removes a blob by its UUID
func (s *PostgresStorage) Delete(ctx context.Context, uuid uuid.UUID) error {
	query := `DELETE FROM signed_blobs WHERE uuid = $1`

	result, err := s.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBlobNotFound
	}

	return nil
}

// PingWithRetry runs a simple query to check if the database is alive, retrying till
func pingWithRetry(
	ctx context.Context, // ctx is the context with timeout for the ping operation
	dataSourceName string, // dataSourceName is the connection string for the database
	retryInterval time.Duration, // retryInterval is the time to wait between retries
) (*sql.DB, error) {

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	var ts time.Time

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("unable to ping database: %w", ctx.Err())
		case <-ticker.C:
			db, err := sql.Open("postgres", dataSourceName)
			if err == nil { // if no error
				// run a simple query to check if the database is alive
				if err = db.QueryRowContext(ctx, selectTimeQuery).Scan(&ts); err == nil {
					return db, err
				}
			}

		}
	}

}

// Ping checks if the storage is reachable
func (s *PostgresStorage) Ping(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, selectTimeQuery); err != nil {
		s.log.Error("failed to ping database", "error", err)
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}
