package pkg

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var (
	verifyDir     string // place to look for blob files, metadata and signatures
	publicKeyPath string // location of the public key on the disk
)

func init() {
	verifyCommand.Flags().StringVar(&publicKeyPath, "public-key", "./public.pem",
		"Path to PEM-encoded public key file (required)")
	verifyCommand.Flags().StringVar(&verifyDir, "dir", ".",
		"Directory to look for blob files (default: current directory)")
	rootCmd.AddCommand(verifyCommand)
}

var verifyCommand = &cobra.Command{
	SilenceUsage: true,
	Use:          "verify <uuid> --public-key <path> --dir <directory-containing-files>",
	Short:        "Verifies the signature of a previously downloaded blob",
	Long: `Verifies the authenticity and integrity of a blob using its signature and metadata.

Requires the public key used by the signing service.

Expected files:
  - <uuid>.txt        : The raw blob content
  - <uuid>.sig        : The base64-encoded signature
  - <uuid>.meta.json  : Metadata with UUID, hash, timestamp

Example:
  ./client verify 10315b7a... --public-key server_pub.pem --directory ./blobs
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("usage: %s verify <uuid> --public-key <path> --dir <directory>", os.Args[0])
		}

		blobUUID := args[0] // uuid of the store blob

		stat, err := os.Stat(verifyDir)
		if err != nil {
			return fmt.Errorf("failed to read %q: %s", storeDir, err)
		}
		if !stat.IsDir() { // not a directory so we can't create all the files
			return fmt.Errorf("%s is not a directory", stat.Name())
		}

		var (
			metaFile string // file containing metadata
			sigFile  string // file containing signature
			blobFile string // file containing the actual data
		)
		// get path for the whole metadata file
		metaFile, err = getAbsolutePath(verifyDir + "/" + blobUUID + ".meta.json")
		if err != nil {
			return fmt.Errorf("unable to read metadata file JSON: %w", err)
		}
		sigFile, err = getAbsolutePath(verifyDir + "/" + blobUUID + ".sig")
		if err != nil {
			return fmt.Errorf("unable to read signature file: %w", err)
		}

		blobFile, err = getAbsolutePath(verifyDir + "/" + blobUUID + ".txt")
		if err != nil {
			return fmt.Errorf("unable to read blob content file: %w", err)
		}

		// Read and then marashl the metadata associated with the blob
		metaBytes, err := os.ReadFile(metaFile)
		if err != nil {
			return fmt.Errorf("failed to read metadata: %w", err)
		}

		var meta metaData

		// since the service signs the metadata bytes
		// we need to do the same
		if err := json.Unmarshal(metaBytes, &meta); err != nil {
			return fmt.Errorf("failed to parse metadata: %w", err)
		}

		// Load the content i.e the blob that was signed
		blobBytes, err := os.ReadFile(blobFile)
		if err != nil {
			return fmt.Errorf("failed to read blob content: %w", err)
		}

		// Load signature
		sigBase64, err := os.ReadFile(sigFile)
		if err != nil {
			return fmt.Errorf("failed to read signature: %w", err)
		}
		sig, err := base64.StdEncoding.DecodeString(string(sigBase64))
		if err != nil {
			return fmt.Errorf("invalid base64 in signature file: %w", err)
		}

		// Compute hash and compare
		hash := sha256.Sum256(blobBytes)
		computedHash := hex.EncodeToString(hash[:])
		if computedHash != meta.Hash {
			return fmt.Errorf("hash mismatch! Expected: %s, Computed: %s", meta.Hash, computedHash)
		}
		log.Printf("✅ Hash matches: %s", computedHash)

		// Rebuild protobuf message
		// this is necesarey because the server signd the byte payload of this
		payload := &blobv1.BlobRecord{
			Uuid:      meta.UUID,
			Blob:      string(blobBytes),
			Hash:      meta.Hash,
			Timestamp: meta.TimeStamp,
		}
		payloadBytes, err := proto.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload for verification: %w", err)
		}

		// we assume that the argument  is a full path
		pubBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read public key: %w", err)
		}
		block, _ := pem.Decode(pubBytes)
		if block == nil || block.Type != "PUBLIC KEY" {
			return fmt.Errorf("invalid PEM format for public key")
		}
		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse public key: %w", err)
		}

		// since we are using RSAPSS to sign, we need to read RSA public key
		rsaPubKey, ok := pubInterface.(*rsa.PublicKey)
		if !ok {
			log.Fatal("public key is not RSA, We use RSAPSS")
		}

		// Verify using RSASSA-PSS
		hashed := sha256.Sum256(payloadBytes)
		err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, hashed[:], sig, &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
			Hash:       crypto.SHA256,
		})
		if err != nil {
			return fmt.Errorf("signature verification failed: %v", err)
		}
		log.Println("✅ Signature verification successful!")

		return nil
	},
}

func getAbsolutePath(fileName string) (string, error) {
	if fileName == "" {
		return "", errors.New("empty filename passed")
	}
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return "", fmt.Errorf("unable to get full path of the filename: %w", err)
	}
	return absPath, nil

}
