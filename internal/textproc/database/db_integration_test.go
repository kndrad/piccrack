//go:build integration

package database

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/kndrad/wcrack/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type postgresContainer struct {
	cfg     *config.Config
	ctr     *postgres.PostgresContainer
	connStr string
}

func newPostgresContainer(ctx context.Context, cfg *config.Config) (*postgresContainer, error) {
	ctr, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithUsername(cfg.Database.User),
		postgres.WithPassword(cfg.Database.Password),
		postgres.WithDatabase(cfg.Database.Name),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	if err != nil {
		return nil, fmt.Errorf("run container: %w", err)
	}

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("connection string: %w", err)
	}

	return &postgresContainer{
		cfg:     cfg,
		ctr:     ctr,
		connStr: connStr,
	}, nil
}

type testFixture struct {
	suite.Suite

	tctr *postgresContainer
	tcfg *testConfig
}

func newFixture(t *testing.T) *testFixture {
	ts := new(testFixture)

	tcfg, err := newTestConfig()
	require.NoError(t, err)

	ts.tcfg = tcfg

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ctr, err := newPostgresContainer(ctx, tcfg.cfg)
	require.NoError(t, err)

	ts.tctr = ctr

	// Apply migrations
	if err := applyMigrations(ctx, ctr.connStr); err != nil {
		t.Fatalf("Failed to apply migrations, err: %v", err)
	}

	return ts
}

func (ts *testFixture) RunCleanup(t *testing.T) {
	if err := testcontainers.TerminateContainer(ts.tctr.ctr); err != nil {
		t.Fatalf("Failed to terminate container, err: %v", err)
	}
	if err := ts.tcfg.Teardown(); err != nil {
		t.Fatalf("Failed to teardown test suite: %v", err)
	}
}

func (ts *testFixture) ContainerConnStr(t *testing.T) string {
	s := ts.tctr.connStr
	require.NotEmpty(t, s)
	return s
}

func TestDatabaseQueries(t *testing.T) {
	t.Parallel()

	fx := newFixture(t)

	require.NotNil(t, fx)

	ctx := context.Background()
	const DefaultQueryLimit int32 = 857
	// Fill the database with words
	t.Run("initial_words_add", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, fx.ContainerConnStr(t))
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := New(conn)
		for _, w := range testWords(t) {
			row, err := q.CreateWord(ctx, w)
			require.NoError(t, err)
			assert.Equal(t, w, row.Value)
		}
	})

	t.Run("list_words", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, fx.ContainerConnStr(t))
		require.NoError(t, err)
		defer conn.Close(ctx)

		query := New(conn)
		params := ListWordsParams{Limit: DefaultQueryLimit}
		rows, err := query.ListWords(ctx, params)
		require.NoError(t, err)
		require.Equal(t, len(testWords(t)), len(rows))
	})

	t.Run("create_word", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, fx.ContainerConnStr(t))
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := New(conn)
		row, err := q.CreateWord(ctx, "test1")
		require.NoError(t, err)
		require.Equal(t, "test1", row.Value)
	})

	t.Run("list_word_frequencies", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, fx.ContainerConnStr(t))
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := New(conn)
		params := ListWordFrequenciesParams{Limit: DefaultQueryLimit}
		rows, err := q.ListWordFrequencies(ctx, params)
		require.NoError(t, err)

		for _, row := range rows {
			// Pick some random words
			switch row.Value {
			case "leading":
				require.Equal(t, int64(2), row.Total)
			case "development":
				require.Equal(t, int64(10), row.Total)
			case "experience":
				require.Equal(t, int64(28), row.Total)
			}
		}
	})

	t.Run("list_word_rankings", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, fx.ContainerConnStr(t))
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := New(conn)
		params := ListWordRankingsParams{Limit: DefaultQueryLimit}
		rows, err := q.ListWordRankings(ctx, params)
		require.NoError(t, err)

		for _, row := range rows {
			// Pick some random words
			switch row.Value {
			case "experience":
				require.Equal(t, int64(1), row.Ranking)
			case "team":
				require.Equal(t, int64(2), row.Ranking)
			case "software":
				require.Equal(t, int64(3), row.Ranking)
			}
		}
	})

	t.Run("list_word_batches", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, fx.ContainerConnStr(t))
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := New(conn)
		params := ListWordBatchesParams{Limit: DefaultQueryLimit}
		batchRows, err := q.ListWordBatches(ctx, params)
		require.NoError(t, err)

		for _, row := range batchRows {
			t.Logf("Got batch: %v", row)
		}
	})

	t.Run("list_words_by_batch_name_returns_batch_ids_and_related_words", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, fx.ContainerConnStr(t))
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := New(conn)
		rows, err := q.ListWordsByBatchName(ctx, "test_batch")
		require.NoError(t, err)

		// Should be 0 for now
		require.Empty(t, rows)

		for _, row := range rows {
			fmt.Printf("LISTING WORDS BY BATCH NAME %v", row)
		}
	})

	fx.RunCleanup(t)
}

func applyMigrations(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	m, err := migrate.New("file://"+filepath.Join("testdata/migrations"), url)
	if err != nil {
		return fmt.Errorf("new migrate: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up migrations: %w", err)
	}

	return nil
}

func testWords(t *testing.T) []string {
	// Fill the database with words
	data, err := os.ReadFile(filepath.Join("testdata", "words.txt"))
	require.NoError(t, err)

	values := make([]string, 0)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		values = append(values, strings.Trim(scanner.Text(), " "))
	}

	if len(values) > math.MaxInt32 {
		t.Fatalf("words len %d exceeds max %d", len(values), math.MaxInt32)
	}

	return values
}

type testConfig struct {
	f   *os.File
	cfg *config.Config
}

func newTestConfig() (*testConfig, error) {
	f, err := os.CreateTemp("testdata", "*.yaml")
	if err != nil {
		return nil, err
	}

	// Write
	content := []byte(`app:
  environment: "testing"

http:
  host: "0.0.0.0"
  port: "8080"
  tls_enabled: false

database:
  user: testuser
  password: testpassword
  host: localhost
  port: 5433
  name: wcrack
  pool:
    max_conns: 25
    min_conns: 5
    max_conn_lifetime: 1h
    max_conn_idle_time: 30m
    connect_timeout: 10s
    dialer_keep_alive: 5s
`)
	if _, err := f.Write(content); err != nil {
		return nil, fmt.Errorf("write data: %w", err)
	}

	cfg, err := config.Load(f.Name())
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &testConfig{
		cfg: cfg,
		f:   f,
	}, nil
}

func (tc *testConfig) Teardown() error {
	if err := os.RemoveAll(tc.f.Name()); err != nil {
		return fmt.Errorf("remove all: %w", err)
	}
	return nil
}
