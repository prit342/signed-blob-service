package e2e

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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
	// define context

	req := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog(logMsg),
		Env: map[string]string{
			"POSTGRES_PASSWORD": dbPass,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_DB":       dbName,
		},
	}

	pgC, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	require.NoError(t, err, "error starting postgres container")
	dbHost, err := pgC.Host(ctx)
	require.NoError(t, err, "error getting the postgres container host")

	mappedPort, err := pgC.MappedPort(ctx, "5432")
	require.NoError(t, err, "error getting the mapped port for postgres container")
	dbPort := mappedPort.Port()
	_, err = strconv.Atoi(dbPort)
	require.NoError(t, err, "error converting port to int")

	return dbHost, dbPort, func() {
		//err := pgC.Terminate(ctx)
		//require.NoError(t, err, "error terminating postgres container")
	}

}
