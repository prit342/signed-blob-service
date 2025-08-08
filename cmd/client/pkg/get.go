package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"
	"github.com/spf13/cobra"
)

var (
	storeDir string // place to strore the downloaded files
)

func init() {
	getCommand.Flags().StringVar(&storeDir, "dir", ".",
		"Directory to store downloaded blob files (default: current directory)")
	rootCmd.AddCommand(getCommand)
}

var getCommand = &cobra.Command{
	Use:          "get <uuid> --dir <place-to-store-files>",
	SilenceUsage: true,
	Short:        "Downloads a signed blob by UUID and saves into various files, including blob content, signature, and metadata",
	Long: `Downloads a signed blob identified by its UUID from the server.

			The following files will be saved:
			- <uuid>.txt     : The raw blob content
			- <uuid>.sig     : The base64-encoded signature
			- <uuid>.meta    : Metadata including UUID, hash, and timestamp

			These files can later be used to verify the integrity and authenticity of the blob.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatal("Please provide a UUID to download")
		}
		blobUUID := args[0]
		if _, err := uuid.Parse(blobUUID); err != nil {
			return fmt.Errorf("invalid UUID format, please provide a valid UUID: %w", err)
		}

		stat, err := os.Stat(storeDir)
		if err != nil {
			return fmt.Errorf("failed to read %q: %s", storeDir, err)
		}
		if !stat.IsDir() { // not a directory so we can't create all the files
			return fmt.Errorf("%s is not a directory", stat.Name())
		}

		resp, err := client.GetSignedBlob(cmd.Context(),
			&blobv1.GetSignedBlobRequest{
				Uuid: blobUUID,
			},
		)

		if err != nil {
			return fmt.Errorf("unable to get blob: %w", err)
		}

		if resp == nil || resp.Payload == nil {
			log.Fatal("response is nil, please check the server logs")
		}

		// write blob contents to to <UUID>.txt
		// in our service we only sign the Payload.Blob, so we can safely write it to a file
		blobFilename := fmt.Sprintf("%s/%s.txt", storeDir, blobUUID)
		if err := os.WriteFile(blobFilename, []byte(resp.Payload.Blob), 0600); err != nil {
			return fmt.Errorf("failed to write blob to file %s: %v", blobFilename, err)
		}

		// write signature to <UUID>.sig (base64-encoded)
		sigFilename := fmt.Sprintf("%s/%s.sig", storeDir, blobUUID)
		sigB64 := base64.StdEncoding.EncodeToString(resp.GetSignature())
		if err := os.WriteFile(sigFilename, []byte(sigB64), 0600); err != nil {
			return fmt.Errorf("failed to write signature to file %q: %w", sigFilename, err)
		}

		// write metadata to <UUID>.meta.json as JSON
		metaFilename := fmt.Sprintf("%s/%s.meta.json", storeDir, blobUUID)

		m := metaData{
			UUID:      resp.GetPayload().GetUuid(),
			Hash:      resp.GetPayload().GetHash(),
			TimeStamp: resp.GetPayload().GetTimestamp(),
		}

		metaByte, err := json.MarshalIndent(&m, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal metadata into JSON: %w", err)
		}

		if err := os.WriteFile(metaFilename, metaByte, 0600); err != nil {
			return fmt.Errorf("faile to write metadata JSON file %s: %w", metaFilename, err)
		}

		// user feedback
		log.Printf("✅ Blob content saved to: %s", blobFilename)
		log.Printf("✅ Signature saved to:    %s", sigFilename)
		log.Printf("ℹ️ Metadata saved to:     %s", metaFilename)

		return nil
	},
}
