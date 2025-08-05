package e2e

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	pgc "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// RunPostgresContainer creates a postgres container
func RunPostgresContainer(
	ctx context.Context,
	t *testing.T,
	postgresImage string,
	logMsg string,
	dbUser string,
	dbPass string,
	dbName string,
) (string, string, func()) {
	t.Helper()

	postgresContainer, err := pgc.Run(ctx,
		"postgres:16-alpine",
		pgc.WithDatabase(dbName),
		pgc.WithUsername(dbUser),
		pgc.WithPassword(dbPass),
	)

	require.NoError(t, err, "error starting postgres container")
	dbHost, err := postgresContainer.Host(ctx)
	require.NoError(t, err, "error getting the postgres container host")

	mappedPort, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err, "error getting the mapped port for postgres container")
	dbPort := mappedPort.Port()
	_, err = strconv.Atoi(dbPort)
	require.NoError(t, err, "error converting port to int")

	return dbHost, dbPort, func() {
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err, "error terminating postgres container")
		t.Logf("Postgres container terminated successfully")
	}

}
