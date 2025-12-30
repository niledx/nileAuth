package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

func TestPostgresIntegration(t *testing.T) {
	if os.Getenv("SKIP_DOCKER") == "1" {
		t.Skip("SKIP_DOCKER=1 set; skipping integration test")
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	// quick ping to ensure daemon reachable
	if err := pool.Client.Ping(); err != nil {
		t.Skipf("docker not available: %v", err)
	}

	// pull postgres and run
	options := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DB=nile_test",
		},
	}
	resource, err := pool.RunWithOptions(options, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err)

	// ensure container is cleaned up
	t.Cleanup(func() {
		_ = pool.Purge(resource)
	})

	var dbURL string
	// exponential backoff-retry to wait for Postgres
	err = pool.Retry(func() error {
		hostPort := resource.GetPort("5432/tcp")
		dbURL = fmt.Sprintf("postgres://test:test@localhost:%s/nile_test?sslmode=disable", hostPort)
		// try to apply migrations which will fail until Postgres is ready
		if err := ApplyMigrations("./migrations", dbURL); err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

	// create Postgres adapter and run basic operations
	pg, err := NewPostgresDB(dbURL)
	require.NoError(t, err)
	defer pg.close()

	// basic user create/get
	u, err := pg.CreateUser("it@example.com", "pwd123", nil)
	require.NoError(t, err)
	require.NotZero(t, u.ID)

	got, err := pg.GetUserByEmail("it@example.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, u.Email, got.Email)

	// refresh token lifecycle
	token := "rt-test-123"
	expires := time.Now().Add(24 * time.Hour).Unix()
	err = pg.CreateRefreshToken(token, u.ID, expires, nil)
	require.NoError(t, err)

	rt, err := pg.GetRefreshToken(token)
	require.NoError(t, err)
	require.NotNil(t, rt)
	require.Equal(t, token, rt.Token)

	// revoke
	err = pg.RevokeRefreshToken(token)
	require.NoError(t, err)

	rt2, err := pg.GetRefreshToken(token)
	require.NoError(t, err)
	require.True(t, rt2.Revoked)

	// revoke all
	err = pg.RevokeAllRefreshTokensForUser(u.ID)
	require.NoError(t, err)

	// ensure ping works
	require.True(t, pg.ping())

	// sanity: ensure environment variable for running tests is present
	_ = os.Setenv("GOTEST_POSTGRES", "1")
}
