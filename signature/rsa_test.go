package signature

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"strings"
	"testing"
)

// helper function to generate a new RSA key pair
// reference -> https://stackoverflow.com/questions/13555085/save-and-load-crypto-rsa-privatekey-to-and-from-the-disk
func generateRsaKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper() // mark this function as a test helper
	privkey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("failed to generate RSA key pair: %v", err)
	}

	return privkey, &privkey.PublicKey
}

func writeRsaPrivateKeyAsPemStringToFile(t *testing.T, privkey *rsa.PrivateKey, filename string) {
	t.Helper() // mark this function as a test helper
	privkeyBytes := x509.MarshalPKCS1PrivateKey(privkey)

	privkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:    "RSA PRIVATE KEY",
			Bytes:   privkeyBytes,
			Headers: map[string]string{},
		},
	)
	if privkeyPem == nil {
		t.Fatalf("failed to encode private key to PEM format")
	}
	err := os.WriteFile(filename, privkeyPem, 0600)
	if err != nil {
		t.Fatalf("failed to write private key to file: %v", err)
	}
}

func TestNewRSASignerServiceFromFile(t *testing.T) {
	// run tests in parallel to speed up execution
	t.Parallel() // run tests in parallel to speed up execution
	tests := []struct {
		name          string
		setupFile     func(t *testing.T) string
		expectError   bool
		errorContains string
	}{
		{
			name: "file exists with valid RSA key",
			setupFile: func(_ *testing.T) string {
				privkey, _ := generateRsaKeyPair(t)
				filename := "test_valid_rsa.pem"
				writeRsaPrivateKeyAsPemStringToFile(t, privkey, filename)

				return filename
			},
			expectError:   false,
			errorContains: "",
		},
		{
			name: "file does not exist",
			setupFile: func(_ *testing.T) string {
				return "nonexistent_file.pem"
			},
			expectError:   true,
			errorContains: "no such file or directory",
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable, not needed in new Go versions

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Setup
			filename := tt.setupFile(t)
			// Cleanup after test
			defer func() {
				if _, err := os.Stat(filename); err == nil {
					_ = os.Remove(filename)
				}
			}()
			// test
			signerService, err := NewRSASignerServiceFromFile(filename)

			if tt.expectError { // if we expect an error
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Fatalf("expected error to contain %q but got %q", tt.errorContains, err.Error())
				}

				return
			}

			// Success case assertions
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if signerService.privateKey == nil || signerService.publicKey == nil {
				t.Fatal("signer service did not initialize keys correctly")
			}

		})
	}
}

func TestRSASignerService_Sign(t *testing.T) {
	// Setup: Create a signer service with test keys
	privkey, _ := generateRsaKeyPair(t)
	signer := &RSASignerService{
		privateKey: privkey,
		publicKey:  &privkey.PublicKey,
	}

	tests := []struct {
		name        string
		blobContent []byte
		expectError bool
	}{
		{
			name:        "sign empty content",
			blobContent: []byte(""),
			expectError: true, // should return error for empty content
		},
		{
			name:        "sign simple text",
			blobContent: []byte("hello world"),
			expectError: false,
		},
		{
			name:        "sign large content",
			blobContent: make([]byte, 10000), // 10KB of zeros
			expectError: false,
		},
		{
			name:        "sign JSON content",
			blobContent: []byte(`{"user":"alice","data":"some blob content"}`),
			expectError: false,
		},
		{
			name:        "sign binary content",
			blobContent: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable, not needed in new Go versions
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			signature, err := signer.Sign(tt.blobContent)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}

				return
			}

			// Success case assertions
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(signature) == 0 {
				t.Fatal("signature should not be empty")
			}

			// Verify that the signature is valid
			err = signer.VerifySignature(tt.blobContent, signature)
			if err != nil {
				t.Fatalf("failed to verify signature: %v", err)
			}

			// appending extra data to the content should invalidate the signature
			// ensure that vefification fails for modified content
			if len(tt.blobContent) > 0 {
				modifiedContent := append(tt.blobContent, byte('x'))
				err := signer.VerifySignature(modifiedContent, signature)
				if err == nil {
					t.Fatal("signature should not be valid for modified content")
				}
			}
		})
	}
}

func TestRSASignerService_SignWithSHA256PKCS1v15_NilPrivateKey(t *testing.T) {
	t.Parallel()
	// Test with nil private key
	signer := &RSASignerService{
		privateKey: nil,
		publicKey:  nil,
	}

	_, err := signer.Sign([]byte("test content"))
	if err == nil {
		t.Fatal("expected error when private key is nil")
	}
}
