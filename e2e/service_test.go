//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	apiv1 "github.com/prit342/signed-blob-service/api/v1"
	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"
	"github.com/prit342/signed-blob-service/logger"
	"github.com/prit342/signed-blob-service/signature"
	"github.com/prit342/signed-blob-service/store"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA2M+2Z3MMViLM1TOUL+wmHSnXPuW2Qm7DbbNJAq7HwL4vql5z
pkbK7Lu2Hc0w9C/uDjJ0zy9+WTRUhpJ8jvoAftUlSckfaWJtzEg2+K+7b05dYcEa
OTn8BBe2+o2SOLELVJul+FyzBOrgeT83XQYFjtgayUIDmAQPF6+Kb9+qkkz6FOUR
mga4VQoYz5+FF+DJGpJl6csPzBItIgNh87uVA6sDTVN/sXJhVuiCodHytgz3sCm0
H8vrq8SbdI70q03Axahr+5R94a99GyyP7KvFp4R/YUwCGduWs1CPNc1nd2Gp4dfX
OIrhI+8zpK2YyetR9fDmdmTx3TNcZM4P7SHlswIDAQABAoIBAADxJtZKjay/5mxN
yXA/lfNdpm3Dh+FVLKh56vB/85Renm1IR94fIA/LKM7wtiCfblz7SdZzTwbmnLNP
rfLdXFMBHxjRlAC1liqt48iOaE2Dx7IKVlNLvsfSy3v8IasQ7HYwzx7F9zfaEL8N
7a2yoIjzRzuhvSStVinGsZFVeNdqEG/vKzT2q1DRp8XChhGJ2+WguswOiEkB5DIq
xBvfszu7rc7ZSN6vS8zHFYtmui6lpzcj7mGqIWNhoHhPcvDayk1xURguwXyAMeZO
h8Uy4hSbudwKkRB1sBdwT4h2Ip/2GBBkN23rDVUajnsxUpvm8Vo2HtoBH/SjyPq5
KmbIgYECgYEA740WMztEp7Iw9a3GqvJxMTkiA+nB2f0kPS3+07ckdAVoEvKJe/AC
s2Jq9ZqxbA0RYmAyXKn4uFp4e9EpKgWV9v9Q4bR4M+nzhyVVJMLqj4Al7E/cihyp
h6wurCQGuLrFtqNG3rroV2lAFb1mXzXo5iKf678V9X9aRvEvrXG5lkMCgYEA57Lm
Dev/8/34owfdGr0LMqT+mD1l5wWXjMu4U77lCoD0MWgQ7/ZPa+cGdff+/MS6xBo0
cnXzP0GKOqqUS6uwRCvL2bgQ1g1OKtM/w853rV6dlt+gEalYtAav5mO94uDtuGPy
5gMj1sg/rENIiEM/Ln0QINLttwmcqr01629X09ECgYEA1+fVpn84tdyI/CWP9etl
0fOokNZS/eKGkw2tq6xZkqh80PcAq0/7XyrJNGwklTqB/KSvP42CusXv6cjuzQ0T
yPb9MzCxVjj6YUhooSV8u7HIfGDOaTzEH6A0wLoHxN+x65bl/UGAv6gBNpbqec3h
B+sVMCmd5RLPjzk6u5zQpHkCgYB68ltRF+IBvsqo+AtDnPzMKvFOJ4ZjSHxaod91
0N4I7NSnQul56+HJCBZNkwMjbeENHjqmYiBpeIW5C7sVTE2EXxkUtq94ZicMYnx7
kpu+y24kGRX/STVgkgvU3Shts51xMtg5ZYEm/6uJ5UofxE9Kg+KDCGpLrjYMA8sQ
20xngQKBgQDCagl7bAUV+o0oChVIwDFUxZwSJ4SlZ0vceHiNRyQhpBmEb48Xn+Q8
OEU2RRrnXmq5ZlSilrxV3P79NCqHVmNkxWu5nwHbhWd/7pdMoeMkPfCIHsQ0bPvN
o7IkTkfGGl5OBBs66IefNY9ZewQor9se5iH1A42DnQ1lh4mtzmB7rg==
-----END RSA PRIVATE KEY-----
`

// Database configuration constants
const (
	// PostgreSQL container configuration
	postgresImage    = "postgres:16-alpine@sha256:dca9c7aa70c71caf8bcc48c8ec40658c1fdf2d1e4e8d26bc27d4d18e030d779a"
	postgresUser     = "testUser"
	postgresPassword = "testPassword"
	postgresDB       = "tesDB"
	postgresPort     = "5432"

	// Application configuration
	appName        = "sign-blob-server"
	appVersion     = "1.0.0"
	appEnvironment = "test"

	// Migration settings
	migrationDir = "../db-migrations/postgres"

	// Test timeouts
	containerStartTimeout = 2 * time.Minute
	testTimeout           = 5 * time.Minute

	postgresContainerReadyMsg = `database system is ready to accept connections`
)

func setupTestDatabase(ctx context.Context, t *testing.T) (store.Storage, func()) {
	// Helper function to set up test database
	t.Helper()

	// spin up a postgres container using testcontainer
	dbHost, dbPort, cleanupFunc := RunPostgresContainer(
		ctx,
		t,
		postgresImage,
		postgresContainerReadyMsg,
		postgresUser,
		postgresPassword,
		postgresDB,
	)

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		postgresUser, postgresPassword, dbHost, dbPort, postgresDB)

	t.Logf("connecting to db: %q", databaseURL)

	log := logger.NewLogger(appName, os.Stdout, slog.LevelDebug, appVersion, appEnvironment)

	// Create storage instance
	storage, err := store.NewPostgresStorage(databaseURL, log, 5*time.Second, testTimeout)
	require.NoError(t, err)

	return storage, cleanupFunc
}

func TestBlobStorageAndVerification(t *testing.T) {
	// This test covers the end-to-end flow of storing a blob, signing it,
	// and then verifying the signature using the public key.
	// It ensures that the blob can be retrieved correctly.

	// Create test context with timeout
	ctxContainer, cancel := context.WithTimeout(context.Background(), containerStartTimeout)
	defer cancel()
	// Set up test database
	storage, cleanup := setupTestDatabase(ctxContainer, t)
	defer cleanup()

	//
	ctx, cancelTest := context.WithTimeout(context.Background(), testTimeout)
	defer cancelTest()
	err := storage.Migrate(ctx, migrationDir)
	require.NoError(t, err, "failed to carry out migrations")

	// Create logger
	log := logger.NewLogger(appName, os.Stdout, slog.LevelDebug, appVersion, appEnvironment)

	// Create temporary file for private key
	privateKeyFile, err := os.CreateTemp("", "private_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(privateKeyFile.Name())

	// Write private key to file
	_, err = privateKeyFile.WriteString(privateKey)
	require.NoError(t, err)
	require.NoError(t, privateKeyFile.Close())

	// Create RSA signer service
	signer, err := signature.NewRSASignerServiceFromFile(privateKeyFile.Name())
	require.NoError(t, err)

	// Initialise the service with storage and signer
	service, err := apiv1.NewService(log, storage, signer)
	require.NoError(t, err)

	// Create a blob to store
	content := "This is a test blob content for end-to-end testing"
	storeReq := &blobv1.StoreBlobRequest{
		Blob: content,
	}

	// Store the blob using service
	t.Log("Storing blob...")
	storeResp, err := service.StoreBlob(ctx, storeReq)
	require.NoError(t, err)
	require.NotEmpty(t, storeResp.Uuid)

	t.Logf("Stored blob with UUID: %s", storeResp.Uuid)

	// Retrieve the signed blob
	getReq := &blobv1.GetSignedBlobRequest{
		Uuid: storeResp.Uuid,
	}
	getResp, err := service.GetSignedBlob(ctx, getReq)
	require.NoError(t, err)
	require.NotNil(t, getResp.Payload)
	require.NotEmpty(t, getResp.Signature)

	// Verify the content matches
	require.Equal(t, content, getResp.Payload.Blob)
	require.Equal(t, storeResp.Uuid, getResp.Payload.Uuid)
	// ensure the hash and timestamp are present
	require.NotEmpty(t, getResp.Payload.Hash)
	require.NotNil(t, getResp.Payload.Timestamp)

	// verify the hash independently, i.e the hash of the content vs the hash in the response
	computedHash := signer.ComputeHash([]byte(content))
	require.NotEmpty(t, computedHash)
	expectedHashStr := hex.EncodeToString(computedHash)
	require.Equal(t, expectedHashStr, getResp.Payload.Hash, "hash mismatch - content integrity compromised")
	t.Logf("Hash verification passed: %s", expectedHashStr)

	//
	// verify the signatures of the payload
	// we reconstruct the exact payload that was signed
	// the UUID, hash, and timestamp
	localPayload := &blobv1.BlobRecord{
		Uuid:      getResp.Payload.Uuid,
		Blob:      content,
		Hash:      expectedHashStr,
		Timestamp: getResp.Payload.Timestamp,
	}
	// marshal the payload exactly as the underlying service logic will do
	b, err := proto.Marshal(localPayload)
	require.NoError(t, err, "failed to marshal payload for verification")

	// PSS is a randomised algorithm — every signature is different, even for the exact same payload and key.
	// so we need to verify the signature for the same content, rather than generating a new signature
	// t.Logf("\n\n[VERIFY] Marshaled payload bytes: %x\n\n", b)
	err = signer.VerifySignature(b, getResp.Signature)
	require.NoError(t, err, "failed to verify signature")

	// Test signature verification fails with tampered content
	tamperedPayload := &blobv1.BlobRecord{
		Uuid:      getResp.Payload.Uuid,
		Blob:      "TAMPERED CONTENT", // Changed content
		Hash:      getResp.Payload.Hash,
		Timestamp: getResp.Payload.Timestamp,
	}
	tamperedSerialised, err := proto.Marshal(tamperedPayload) // Marshal the original payload for signature verification
	require.NoError(t, err)

	err = signer.VerifySignature(tamperedSerialised, getResp.Signature)
	require.Error(t, err, "signature verification should fail for tampered content")
	t.Log("Tamper detection working correctly")

	// Test with wrong signature
	wrongSignature := make([]byte, len(getResp.Signature))
	copy(wrongSignature, getResp.Signature)
	wrongSignature[0] ^= 0xFF // Flip bits in first byte

	err = signer.VerifySignature(b, wrongSignature)
	require.Error(t, err, "verification should fail with wrong signature")
	t.Log("Wrong signature detection working correctly")

}

// TestEdgeCases tests various edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	ctxContainer, cancel := context.WithTimeout(context.Background(), containerStartTimeout)
	defer cancel()
	storage, cleanup := setupTestDatabase(ctxContainer, t)
	defer cleanup()

	ctx, cancelTest := context.WithTimeout(context.Background(), testTimeout)
	defer cancelTest()
	err := storage.Migrate(ctx, migrationDir)
	require.NoError(t, err)

	log := logger.NewLogger(appName, os.Stdout, slog.LevelDebug, appVersion, appEnvironment)

	privateKeyFile, err := os.CreateTemp("", "private_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(privateKeyFile.Name())
	_, err = privateKeyFile.WriteString(privateKey)
	require.NoError(t, err)
	require.NoError(t, privateKeyFile.Close())

	signer, err := signature.NewRSASignerServiceFromFile(privateKeyFile.Name())
	require.NoError(t, err)
	service, err := apiv1.NewService(log, storage, signer)
	require.NoError(t, err)

	// Test empty content
	t.Run("EmptyContent", func(t *testing.T) {
		_, err := service.StoreBlob(ctx, &blobv1.StoreBlobRequest{Blob: ""})
		require.Error(t, err, "should reject empty content")
	})

	// Test nil request
	t.Run("NilRequest", func(t *testing.T) {
		_, err := service.StoreBlob(ctx, nil)
		require.Error(t, err, "should reject nil request")
	})

	// Test invalid UUID for retrieval
	t.Run("InvalidUUID", func(t *testing.T) {
		_, err := service.GetSignedBlob(ctx, &blobv1.GetSignedBlobRequest{Uuid: "invalid-uuid"})
		require.Error(t, err, "should reject invalid UUID")
	})

	// Test non-existent UUID
	t.Run("NonExistentUUID", func(t *testing.T) {
		_, err := service.GetSignedBlob(ctx, &blobv1.GetSignedBlobRequest{Uuid: "550e8400-e29b-41d4-a716-446655440000"})
		require.Error(t, err, "should return error for non-existent UUID")
	})

	// Test max content allowed
	t.Run("LargeContent within limits", func(t *testing.T) {
		largeContent := string(bytes.Repeat([]byte("A"), 256*1024)) // 256 KB of 'A'
		resp, err := service.StoreBlob(ctx, &blobv1.StoreBlobRequest{Blob: largeContent})
		require.NoError(t, err, "should handle large content")

		getResp, err := service.GetSignedBlob(ctx, &blobv1.GetSignedBlobRequest{Uuid: resp.Uuid})
		require.NoError(t, err)
		require.Equal(t, largeContent, getResp.Payload.Blob, "large content should be preserved")
	})

	t.Run("content size is larger than allowed", func(t *testing.T) {
		largeContent := string(bytes.Repeat([]byte("A"), 1*1024*1024)) // 3MB of 'A'

		_, err := service.StoreBlob(ctx, &blobv1.StoreBlobRequest{Blob: largeContent})
		require.Error(t, err, "should reject content larger than 256KB")
	})

}
