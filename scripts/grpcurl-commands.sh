#!/bin/bash
set -euo pipefail

## we are not using reflection, due to security reasons
echo "Testing gRPC service with grpcurl using proto files..."

# List all services using proto files
echo "Listing all services:"
grpcurl -plaintext -import-path ./proto -proto blob/v1/blob.proto localhost:55555 list

# List all methods for BlobService using proto files
echo "Listing methods for BlobService:"
grpcurl -plaintext -import-path ./proto -proto blob/v1/blob.proto localhost:55555 list blob.v1.BlobService

# Describe the StoreBlob method using proto files
echo "Describing StoreBlob method:"
grpcurl -plaintext -import-path ./proto -proto blob/v1/blob.proto localhost:55555 describe blob.v1.BlobService.StoreBlob

# Describe the GetSignedBlob method using proto files
echo "Describing GetSignedBlob method:"
grpcurl -plaintext -import-path ./proto -proto blob/v1/blob.proto localhost:55555 describe blob.v1.BlobService.GetSignedBlob

# # Example call to StoreBlob using proto files
echo "Calling StoreBlob with test data:"
grpcurl -plaintext -import-path ./proto -proto blob/v1/blob.proto \
  -d '{"blob": "SGVsbG8gd29ybGQ="}' \
  localhost:55555 blob.v1.BlobService/StoreBlob

# # Example call to GetSignedBlob using proto files (replace UUID with actual one from StoreBlob response)
# echo "Example GetSignedBlob call (replace UUID):"
echo "grpcurl -plaintext -import-path ./proto -proto blob/v1/blob.proto \\"
echo "  -d '{\"uuid\": \"your-uuid-here\"}' \\"
echo "  localhost:55555 blob.v1.BlobService/GetSignedBlob"
