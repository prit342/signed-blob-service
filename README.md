# 🛡️ Signed Blob Storage Service

A complete blob storage system built using **Go**, **gRPC**, **Protocol Buffers**, and **PostgreSQL**. This service stores text blobs in a cryptographically signed format using RSA-PSS digital signatures, ensuring data integrity, authenticity, and non-repudiation.

The system provides both server-side storage and a command-line client for seamless interaction. When a blob is uploaded, the server generates a unique UUID and creates cryptographic signatures. Clients can then download the original content along with verification files including:

- **`<uuid>.txt`** - The original blob content
- **`<uuid>.sig`** - Base64-encoded RSA-PSS signature  
- **`<uuid>.meta.json`** - Metadata with UUID, SHA-256 hash, and timestamp

This architecture enables secure, verifiable blob storage with complete offline verification capabilities using industry-standard cryptographic methods.

## ✨ Features

- **Secure Storage**: Store plain text blobs with automatic server-side digital signing
- **RSA-PSS Signatures**: Modern probabilistic signature scheme using 2048-bit RSA keys
- **Integrity Verification**: SHA-256 hashing ensures blob content has not been tampered with
- **Client-Side Verification**: Retrieve public key to verify signatures independently
- **UUID-Based Lookup**: Globally unique identifiers for efficient blob retrieval
- **Size Limits**: Configurable blob size limits (currently 256KB maximum)
- **Clean Architecture**: Well-structured codebase with proper separation of concerns
- **Comprehensive Testing**: Unit tests, integration tests using testcontainers, and end-to-end cryptographic verification


## 🧱 Tech Stack

### Backend
- **Go 1.22+**: Modern, performant backend language
- **gRPC**: High-performance RPC framework with HTTP/2 support
- **Protocol Buffers**: Efficient serialisation with strong typing via [Buf](https://buf.build/)
- **PostgreSQL**: Robust relational database for blob storage
- **RSA Cryptography**: Industry-standard digital signatures (RSASSA-PSS, PKCS#1 v2.1)
### Development & Operations
- **Docker Compose**: Multi-service development environment
- **Testcontainers**: Isolated integration testing with real PostgreSQL instances
- **Structured Logging**: JSON-formatted logs with source file information (`slog`)
- **Database Migrations**: Version-controlled schema evolution and migrations using go-migrate


## 📁 Project Structure

| Directory | Purpose | Description |
|-----------|---------|-------------|
| `api/` | gRPC Service Implementation | Houses the main gRPC API service handlers and business logic |
| `gen/` | Generated Protocol Buffer Code | Contains compiled Protocol Buffer definitions and gRPC service stubs |
| `cmd/` | Application Entry Points | Contains main applications (server and client executables) |
| `cmd/client/` | CLI Client Application | Command-line interface for interacting with the blob service |
| `cmd/server/` | gRPC Server Application | Main server application that hosts the blob storage service |
| `db-migrations/` | Database Schema | SQL migration files for PostgreSQL schema versioning used by go-migrate |
| `e2e/` | Integration Tests | End-to-end tests using testcontainers with real database instances |
| `internal/` | Private Application Code | Internal packages not meant for external import |
| `internal/api/` | Service Implementation | gRPC service handlers and business logic |
| `internal/store/` | Data Persistence | Database access layer and storage abstractions |
| `proto/` | Protocol Buffer Definitions | Source `.proto` files defining the gRPC service interface |
| `scripts/` | Development Scripts | Shell scripts for key generation, setup, and development tasks |
| `signature/` | Cryptographic Operations | RSA-PSS signing and verification implementation |

### Key Configuration Files

| File | Purpose |
|------|---------|
| `buf.yaml` | Buf CLI configuration for Protocol Buffer management |
| `buf.gen.yaml` | Code generation settings for Protocol Buffers |
| `docker-compose.yaml` | Multi-service development environment (PostgreSQL + App) |
| `Dockerfile` | Container image definition for the gRPC server |
| `env-local-sample` | Environment variable template for local development |
| `.golangci.yml` | Go linting configuration with project-specific rules |
| `Makefile` | Build automation and development workflow commands |

---

## 📦 gRPC API Reference

The service exposes three core RPC methods:

| Method | Purpose | Input | Output |
|--------|---------|-------|--------|
| `StoreBlob` | Upload and sign a text blob | `StoreBlobRequest` | `StoreBlobResponse` |
| `GetSignedBlob` | Retrieve signed blob with signature | `GetSignedBlobRequest` | `GetSignedBlobResponse` |
| `GetPublicKey` | Fetch server's public signing key | `GetPublicKeyRequest` | `GetPublicKeyResponse` |

### Message Structures

#### `BlobRecord` (Canonical Signed Structure)
```protobuf
message BlobRecord {
  string uuid = 1;      // Server-generated UUID for identification
  string blob = 2;      // Original user-submitted text blob  
  string hash = 3;      // SHA-256 hash of the blob, hex-encoded
  string timestamp = 4; // RFC3339 formatted timestamp (e.g., "2025-07-30T16:52:13Z")
}
```

#### Key API Flows

1. **Store Blob**: Client sends raw text → Server generates UUID, computes hash, adds timestamp → Signs entire `BlobRecord` → Stores in database
2. **Retrieve Blob**: Client requests by UUID → Server returns original `BlobRecord` + RSA signature
3. **Verify Signature**: Client can verify the signature using the public key to ensure data integrity


## 🔐 Cryptographic Details

### Digital Signature Algorithm
- **Algorithm**: RSA-PSS (RSASSA-PSS) - Probabilistic Signature Scheme
- **Key Size**: 2048 bits
- **Hash Function**: SHA-256
- **Salt Length**: Equal to hash length (32 bytes for SHA-256)
- **Signed Data**: Protobuf-serialised `BlobRecord` (includes UUID, blob, hash, timestamp)

### Security Properties
- **Authenticity**: Signatures prove the blob was signed by the server's private key
- **Integrity**: Any tampering with the blob content invalidates the signature
- **Non-repudiation**: Server cannot deny having signed a blob
- **Probabilistic**: Each signature includes random salt, making signatures non-deterministic (enhanced security)
- **Provable Security**: RSA-PSS has stronger security proofs than PKCS#1 v1.5

### Key Management
- Private key stored securely on server (never transmitted)
- Public key available via gRPC endpoint for client verification
- Keys generated in PKCS#1 format for compatibility


## 🛠️ Development Scripts

The project includes several utility scripts and Make targets for development and deployment:

### Make Targets

The Makefile provides comprehensive build and development targets with built-in help:

```bash
make help                 # Display all available targets with descriptions
make                      # Same as 'make help' (default target)
```

**Key targets organised by category:**

```bash
# Build & Development
make generate-proto       # Generate Go code from Protocol Buffer definitions
make build                # Build Docker images with generated code and RSA keys  
make run                  # Start the service stack (requires prior build)
make build-and-run        # Build and start the complete service in one command

# Testing & Quality Assurance  
make unit-test            # Run unit tests for all Go packages
make e2e-test             # Run end-to-end integration tests using testcontainers
make test                 # Run complete test suite (unit + integration tests)

# Client Application
make build-client         # Build the command-line client application

# Maintenance
make clean                # Remove generated files and Docker resources
```

### Shell Scripts

```bash
./generate-rsa-keys.sh    # Generate 2048-bit RSA private key in PKCS#1 format
```
- Creates `private_key.pem` with proper permissions (644)
- Converts from PKCS#8 to PKCS#1 format for Go compatibility
- Validates key format before completion

### Build & run
```bash
make build-and-run        # One-step build and run with latest changes
```

### Testing
```bash
make unit-test            # run all unit tests
make e2e-test             # run end-to-end test using testcontainers
make test                 # Run all tests
```

### Testing the API

- Run the gRPC server in one terminal:
```
make check
make build-and-run
```

- Build the client in another terminal:
```bash
make build-client
```

- Create a file with some contents:
```bash
echo hello-world > /tmp/test.txt
```

- Upload/Put the file:

```
❯ ./client --server localhost:55555 put /tmp/test.txt
2025/08/02 11:37:49 Blob stored successfully with UUID: 9de22b2a-9d35-42d8-8b7e-fd2570aca13b
```

- Download the same content along with signature to a local folder download:

```bash
❯ mkdir downloads  
❯ ./client --server localhost:55555 get 9de22b2a-9d35-42d8-8b7e-fd2570aca13b --dir ./downloads 
2025/08/02 11:40:44 ✅ Blob content saved to: ./downloads/9de22b2a-9d35-42d8-8b7e-fd2570aca13b.txt
2025/08/02 11:40:44 ✅ Signature saved to:    ./downloads/9de22b2a-9d35-42d8-8b7e-fd2570aca13b.sig
2025/08/02 11:40:44 ℹ️ Metadata saved to:     ./downloads/9de22b2a-9d35-42d8-8b7e-fd2570aca13b.meta.json

```
- Download public key from the server:
```bash
❯ ./client --server localhost:55555 get-public-key public.pem
2025/08/02 11:41:50 ✅ Public key saved to file: public.pem

```

- Verify the downloaded files using the local public key (offline verification):

```bash
./client verify 9de22b2a-9d35-42d8-8b7e-fd2570aca13b --dir downloads --public-key public.pem
2025/08/02 11:43:37 ✅ Hash matches: d79f2e37784e5cd8631963896ebc6c9c66934af94a1854504717eaec04bc3d09
2025/08/02 11:43:37 ✅ Signature verification successful!
```

## 🔧 Configuration
- All the environment variables are documented in the sample `env-local-sample` file that can be used for testing.


## License
MIT License


## 🚀 Future Improvements

- **Tracing**: Add OpenTelemetry support for tracing and metrics
- **Metrics Collection**: Implement Prometheus metrics for blob operations, signature performance, and database queries
- **Health Checks**: Enhanced health endpoints with dependency status (database, key availability)
- **External Key Management**: Integration with AWS KMS, Azure Key Vault, and HashiCorp Vault
- **TLS Mutual Authentication**: Client certificate validation for enhanced security
- **Caching Layer**: Redis integration for frequently accessed blobs and public keys
- **REST API**: Optional HTTP REST interface alongside gRPC for broader client compatibility
- **Go Tool**: Use `go tool` for installing various command line utilities and migrate to go 1.24
