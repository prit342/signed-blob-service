package pkg

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(putCommand)
}

var putCommand = &cobra.Command{
	Use:          "put <filename>",
	SilenceUsage: true,
	Short:        "uploads a blob from a file and then and return its unique UUID",
	Long:         `uploads a blob of content to the Sign-Blob-Service and return its UUID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Please provide a file name to upload")
		}
		filename := args[0]
		if filename == "" {
			return errors.New("Please provide a file name to upload")
		}
		// grab the full path of the file
		fullPath, err := filepath.Abs(filename)
		if err != nil {
			return fmt.Errorf("unable to get full path of the file: %w", err)
		}

		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", fullPath, err)
		}

		if fileInfo.IsDir() {
			return fmt.Errorf("%q is a directory, please provide a file", fullPath)
		}
		// check if the fileInfo is a regular file and not a
		if !fileInfo.Mode().IsRegular() {
			return fmt.Errorf("file %s is not a regular file", fullPath)
		}

		// open the file and read its content
		file, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("error opening file %s: %w", filename, err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("error closing file %s: %s", filename, err)
			}
		}()

		b, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", filename, err)
		}

		if len(b) == 0 {
			return fmt.Errorf("file is %q empty, please provide a file with content", fullPath)
		}

		resp, err := client.StoreBlob(cmd.Context(), &blobv1.StoreBlobRequest{
			Blob: string(b),
		})

		if err != nil {
			return fmt.Errorf("error storing blob: %w", err)
		}

		if resp == nil {
			return errors.New("got empty response from the server")
		}

		log.Printf("Blob stored successfully with UUID: %s", resp.GetUuid())

		return nil
	},
}
