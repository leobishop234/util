package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/leobishop234/util/srverr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	psqlContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDatabaseImage   = "postgres:17-alpine"
	testDatabaseName    = "test"
	testDatabaseUser    = "test"
	testDatabasePass    = "test"
	testDatabasePort    = "5432"
	testDatabaseSslmode = "disable"
)

type testDBContainer struct {
	*psqlContainer.PostgresContainer
}

type Closer interface {
	Close() error
}

type TestDB[D Closer] struct {
	db                D
	postgresContainer *testDBContainer
	devContainer      *testDBContainer
	closeOnce         sync.Once
	closeErr          error
}

func newTestDBConfig(ctx context.Context) (DBConfig, *testDBContainer, error) {
	postgresContainer, err := psqlContainer.Run(ctx,
		testDatabaseImage,
		psqlContainer.WithDatabase(testDatabaseName),
		psqlContainer.WithUsername(testDatabaseUser),
		psqlContainer.WithPassword(testDatabasePass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		return DBConfig{}, nil, Error("failed to start postgres test container", err, srverr.ErrCodeDependencyFailure)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return DBConfig{}, nil, Error("failed to get postgres test container host", err, srverr.ErrCodeDependencyFailure)
	}

	port, err := postgresContainer.MappedPort(ctx, testDatabasePort)
	if err != nil {
		return DBConfig{}, nil, Error("failed to get postgres test container port", err, srverr.ErrCodeDependencyFailure)
	}

	// Use 127.0.0.1 if host resolves to localhost to avoid IPv6 issues.
	if host == "localhost" {
		host = "127.0.0.1"
	}

	return DBConfig{
		Host:     host,
		Port:     int(port.Num()),
		Database: testDatabaseName,
		User:     testDatabaseUser,
		Pass:     testDatabasePass,
		Sslmode:  testDatabaseSslmode,
	}, &testDBContainer{PostgresContainer: postgresContainer}, nil
}

func NewServiceTestDB[D Closer](
	t *testing.T,
	open func(ctx context.Context, conf, dev DBConfig) (D, error),
	seed func(t *testing.T, db D),
) *TestDB[D] {
	t.Helper()

	config, container, err := newTestDBConfig(t.Context())
	require.NoError(t, err)

	devConfig, devContainer, err := newTestDBConfig(t.Context())
	require.NoError(t, err)

	db, err := open(t.Context(), config, devConfig)
	require.NoError(t, err)

	if seed != nil {
		seed(t, db)
	}

	testDB := &TestDB[D]{
		db:                db,
		postgresContainer: container,
		devContainer:      devContainer,
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), 30*time.Second)
		defer cancel()
		assert.NoError(t, testDB.Close(cleanupCtx))
	})

	return testDB
}

func (t *TestDB[D]) DB() D {
	return t.db
}

func (t *TestDB[D]) Close(ctx context.Context) error {
	t.closeOnce.Do(func() {
		if ctx == nil {
			ctx = context.Background()
		}

		var closeErr error
		if err := t.db.Close(); err != nil {
			closeErr = errors.Join(closeErr, Error("failed to close service test db", err, srverr.ErrCodeDependencyFailure))
		}

		closeErr = errors.Join(closeErr, terminateTestDB(ctx, t.postgresContainer))
		closeErr = errors.Join(closeErr, terminateTestDB(ctx, t.devContainer))
		t.closeErr = closeErr
	})

	return t.closeErr
}

func terminateTestDB(ctx context.Context, container *testDBContainer) error {
	if container == nil {
		return nil
	}

	if err := container.Terminate(ctx); err != nil {
		return Error("failed to terminate postgres test container", err, srverr.ErrCodeDependencyFailure)
	}

	return nil
}
