package pkg

import (
	"context"
	"log"

	blobv1 "github.com/prit342/signed-blob-service/gen/blob/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	server string
	conn   *grpc.ClientConn
	client blobv1.BlobServiceClient
)

var rootCmd = &cobra.Command{
	Use:   "sign-blob-client",
	Short: "A simple client for the Signed Blob Service",
	Long:  `A simple client for the Signed Blob Service that allows you to store and retrieve signed blobs.`,

	// PersistentPreRun in the root command to initialize the client
	// after flags are parsed but before any subcommand runs
	// this allows us to set up the client connection once
	// and reuse it across all subcommands
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		InitClient()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		CloseClient() // close the client connection after all commands are done
	},
}

// InitClient initializes the gRPC client connection to the server.
func InitClient() {
	var err error
	conn, err = grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	client = blobv1.NewBlobServiceClient(conn)
}

// CloseClient closes the gRPC client connection.
func CloseClient() {
	if conn != nil {
		_ = conn.Close()
	}
}

// ExecuteWithContext - run the root command with context.
func ExecuteWithContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&server, "server", "localhost:12345", "address of the GRPC server to connect to")
}
