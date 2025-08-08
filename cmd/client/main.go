package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/prit342/signed-blob-service/cmd/client/pkg"
	"golang.org/x/sync/errgroup"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel() // ensure we stop the context to free resources

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return pkg.ExecuteWithContext(ctx)
	})

	if err := g.Wait(); err != nil {
		cancel()
		log.Fatalf("Error running application: %v", err)
	}
}
