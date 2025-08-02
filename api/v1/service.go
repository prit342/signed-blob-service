package v1

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"
	"github.com/prit342/signed-blob-service/signature"
	"github.com/prit342/signed-blob-service/store"
	"google.golang.org/protobuf/proto"
)

// Sever represents the main server structure
type Service struct {
	blobv1.UnimplementedBlobServiceServer // Embed the generated server interface
	logger                                *slog.Logger
	store                                 store.Storage
	signer                                signature.Signer
}

// we only allow blobs of size 256 Kilobytes
const maxBlobSize = 256 * 1024 // 256KB in bytes

// NewServer creates a new instance of Sever with the provided dependencies
func NewService(logger *slog.Logger, storage store.Storage, signer signature.Signer) (*Service, error) {
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}
	if storage == nil {
		return nil, errors.New("storage cannot be nil")
	}
	if signer == nil {
		return nil, errors.New("signer cannot be nil")
	}
	return &Service{
		logger: logger,
		store:  storage,
		signer: signer,
	}, nil
}

// StoreBlob stores a blob and its signature and returns its UUID
func (s *Service) StoreBlob(ctx context.Context, req *blobv1.StoreBlobRequest) (*blobv1.StoreBlobResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	if req.Blob == "" {
		return nil, errors.New("blob content cannot be empty")
	}

	// we reject blobs larger that 2 MB for the time being
	if len(req.Blob) > maxBlobSize {
		return nil, fmt.Errorf("blob content exceeds maximum size of %d bytes", maxBlobSize)
	}

	hash := s.signer.ComputeHash([]byte(req.Blob))
	if len(hash) == 0 {
		s.logger.Error("failed to compute hash for blob content")
		return nil, errors.New("failed to compute hash for blob content")
	}

	// Encode the hash to a string for storage
	// This is necessary because the storage expects a string representation of the hash
	encodedHashStr := hex.EncodeToString(hash[:])

	uuidStr := uuid.New().String() // the uuid for the blob
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// this is the payload we will sign
	payloadToBeSigned := &blobv1.BlobRecord{
		Uuid:      uuidStr,
		Blob:      req.Blob,
		Hash:      encodedHashStr,
		Timestamp: timestamp,
	}

	// we need to marshal the payload to bytes before signing
	// this is because the signer expects a byte slice to sign
	serialisedPayload, err := proto.Marshal(payloadToBeSigned)
	s.logger.Debug("[SIGN] Marshaled payload bytes", "bytes",
		fmt.Sprintf("%x", serialisedPayload))

	if err != nil {
		s.logger.Error("failed to marshal payload", "error", err)
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	// instead of signing just the content, we sign the entire request
	// this ensures that the signature is valid for the entire request structure

	sig, err := s.signer.Sign(serialisedPayload)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to sign the payload: %v", err))
		return nil, fmt.Errorf("failed to sign payload: %w", err)
	}

	// Create a new Blob instance to store
	recordWithSignature := &blobv1.SignedBlobRecord{
		Payload: &blobv1.BlobRecord{
			Uuid:      uuidStr,
			Blob:      req.Blob,
			Hash:      encodedHashStr,
			Timestamp: timestamp,
		},
		Signature: sig,
	}

	// fmt.Printf("\n%+v\n", recordBlob)
	//s.logger.Debug("recording blob", "blob", fmt.Sprintf("%x", recordWithSignature))

	if err := s.store.Store(ctx, recordWithSignature); err != nil {
		s.logger.Error(fmt.Sprintf("failed to store signed recored: %v", err))
		return nil, fmt.Errorf("failed to store signed record: %w", err)
	}

	return &blobv1.StoreBlobResponse{
		Uuid: uuidStr,
	}, nil
}

// GetSignedBlob retrieves a signed blob by its UUID
func (s *Service) GetSignedBlob(ctx context.Context, req *blobv1.GetSignedBlobRequest) (*blobv1.GetSignedBlobResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	if req.Uuid == "" {
		return nil, errors.New("UUID cannot be empty")
	}

	uuid, err := uuid.Parse(req.Uuid)
	if err != nil {
		s.logger.Error("failed to parse UUID", "error", err)
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	// grab the blob from the storage
	blobRow, err := s.store.GetByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, store.ErrBlobNotFound) {
			return nil, fmt.Errorf("blob not found: %w", err)
		}
		s.logger.Error(fmt.Sprintf("failed to retrieve blob: %v", err))
		return nil, fmt.Errorf("failed to retrieve blob: %w", err)
	}

	signature := blobRow.Signature
	if len(signature) == 0 {
		return nil, errors.New("signature is empty")
	}

	response := &blobv1.GetSignedBlobResponse{
		Payload: &blobv1.BlobRecord{
			Uuid:      blobRow.Payload.Uuid,
			Hash:      blobRow.Payload.Hash,
			Blob:      blobRow.Payload.Blob,
			Timestamp: blobRow.Payload.Timestamp,
		},
		Signature: signature,
	}

	return response, nil
}

// GetPublicKey returns the public key used for signing blobs
func (s *Service) GetPublicKey(context.Context, *blobv1.GetPublicKeyRequest) (*blobv1.GetPublicKeyResponse, error) {
	if s.signer == nil {
		return nil, errors.New("signer is not initialized")
	}
	publicKey, err := s.signer.GetPublicKey()
	if err != nil {
		s.logger.Error("failed to retrieve public key", "error", err)
		return nil, fmt.Errorf("failed to retrieve public key: %w", err)
	}
	return &blobv1.GetPublicKeyResponse{
		PublicKey: string(publicKey),
	}, nil

}
