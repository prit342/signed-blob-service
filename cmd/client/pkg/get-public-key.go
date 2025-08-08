package pkg

import (
	"fmt"
	"log"
	"os"

	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getPublicKeyCommand)
}

var getPublicKeyCommand = &cobra.Command{
	Use:          "get-public-key <uuid>",
	SilenceUsage: true,
	Short:        "Downloads the public key associated with a signed blob service and stores it in a file.",
	Long: `Fetches the public key used by the signed blob server and saves it locally.

		   Overrides the destination file if it already exists but does not change
		   The public key can be used to verify the authenticity of signed blobs offline.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatalf("Usage: %s get-public-key <filename>\nPlease provide a filename to save the public key.", os.Args[0])
		}

		publicKeyFile := args[0]

		resp, err := client.GetPublicKey(cmd.Context(),
			&blobv1.GetPublicKeyRequest{},
		)

		if err != nil {
			return fmt.Errorf("unable to write file %q: %w", publicKeyFile, err)
		}

		// write the public key to the file
		if err := os.WriteFile(publicKeyFile, []byte(resp.PublicKey), 0600); err != nil {
			return fmt.Errorf("failed to write blob to file %s: %w", publicKeyFile, err)
		}
		// user feedback
		log.Printf("✅ Public key saved to file: %s", publicKeyFile)

		return nil
	},
}
