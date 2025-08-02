package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// RSASignerService handles signing and verifying content using RSA keys.
type RSASignerService struct {
	privateKey *rsa.PrivateKey // Server's private key (used for signing)
	publicKey  *rsa.PublicKey  // Server's public key (used for verification)
}

var _ Signer = (*RSASignerService)(nil)

// rsaPSSOptions - contains options for creating and verifying PSS signatures using RSA
var rsaPSSOptions = rsa.PSSOptions{
	// Use a salt length equal to the length of the hash function output (32 bytes for SHA-256)
	// This is the most common and interoperable setting.
	SaltLength: rsa.PSSSaltLengthEqualsHash,
	Hash:       crypto.SHA256,
}

// NewRSASignerServiceFromFile loads a PEM-encoded RSA private key from a file,
// derives the corresponding public key, and returns a signer service.
func NewRSASignerServiceFromFile(pemFile string) (*RSASignerService, error) {
	// Read the private key file
	keyBytes, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// Decode the PEM block
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing RSA private key")
	}

	// Parse the private key from the PEM block (expects PKCS#1 format)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	// Derive the public key from the private key
	publicKey := &privateKey.PublicKey

	return &RSASignerService{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// Sign - signs the input payload using RSASSA-PSS with SHA-256.
// This is a probabilistic signature scheme (includes random salt for every signature).
func (s *RSASignerService) Sign(blobContent []byte) ([]byte, error) {

	if err := rSASignerServiceCheckInit(s); err != nil {
		return nil, fmt.Errorf("failed to sign content: %w", err)
	}

	if len(blobContent) == 0 {
		return nil, errors.New("blob content cannot be nil or empty")
	}

	// Step 1: computer the Hash the input (SHA-256)
	hashed := sha256.Sum256(blobContent)

	// Sign the blobContent using the private key and PSS scheme
	signature, err := rsa.SignPSS(rand.Reader, s.privateKey, cryptoHash(), hashed[:], &rsaPSSOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to sign blob content using RSASSA-PSS: %w", err)
	}

	return signature, nil
}

// ComputeHash - computes the SHA-256 hash of the given blob content.
func (s *RSASignerService) ComputeHash(blobContent []byte) []byte {
	// Compute the SHA-256 hash
	hash := sha256.Sum256(blobContent)
	return hash[:]
}

// VerifySignature - checks whether the given signature is valid for the provided blobContent
// using the server's RSA public key.
func (s *RSASignerService) VerifySignature(blobContent []byte, signature []byte) error {

	if err := rSASignerServiceCheckInit(s); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	if len(blobContent) == 0 {
		return errors.New("blob content cannot be nil or empty")
	}

	if len(signature) == 0 {
		return errors.New("signature content cannot be nil or empty")
	}
	// Hash the content using the same algorithm
	hashed := sha256.Sum256(blobContent)

	// Verify the signature using the public key
	err := rsa.VerifyPSS(s.publicKey, cryptoHash(), hashed[:], signature, &rsaPSSOptions)
	if err != nil {
		return fmt.Errorf("signature verification failed using RSASSA-PSS: %w", err)
	}

	return nil
}

// rSASignerServiceCheckInit - checks to see if the RSASigner service is initialised properly
func rSASignerServiceCheckInit(s *RSASignerService) error {
	if s == nil {
		return errors.New("RSASignerService has not been initialised properly")
	}
	if s.privateKey == nil || s.publicKey == nil {
		return errors.New("signer service is not properly initialised with keys")
	}
	return nil
}

// GetPublicKey returns the PEM-encoded public key in PKIX format.
// This can be safely shared with clients for signature verification.
func (s *RSASignerService) GetPublicKey() ([]byte, error) {

	if s.publicKey == nil {
		return nil, errors.New("signer service is not properly initialised with keys")
	}
	// Marshal the public key to ASN.1 DER-encoded PKIX format
	pubASN1, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Encode it to PEM format
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubPEM, nil
}

// getPrivateKey returns the PEM-encoded private key in PKCS#1 format.
// This is used internally and never exposed in production environments.
func (s *RSASignerService) getPrivateKey() ([]byte, error) {
	// Marshal the private key to ASN.1 DER-encoded PKCS#1 format
	privASN1 := x509.MarshalPKCS1PrivateKey(s.privateKey)

	// Encode it to PEM format
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privASN1,
	})

	return privPEM, nil
}

// cryptoHash returns the hash function used throughout the service (SHA-256).
// RSA with PKCS1v15 signing requires specifying the hash algorithm.
func cryptoHash() crypto.Hash {
	return crypto.SHA256
}
