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
	"github.com/kndrad/piccrack/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type DatabaseTestSuite struct {
	suite.Suite
	container *postgres.PostgresContainer
	connStr   string
	cfg       *config.Config
	cfgPath   string
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupSuite() {
	tcfg, err := newTestConfig()
	require.NoError(s.T(), err)
	s.cfg = tcfg.cfg
	s.cfgPath = tcfg.f.Name()

	ctx := context.Background()
	container, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithUsername(s.cfg.Database.User),
		postgres.WithPassword(s.cfg.Database.Password),
		postgres.WithDatabase(s.cfg.Database.Name),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	require.NoError(s.T(), err)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	s.container = container
	s.connStr = connStr

	// Apply migrations
	err = applyMigrations(ctx, s.connStr)
	require.NoError(s.T(), err)
}

func (s *DatabaseTestSuite) TearDownSuite() {
	ctx := context.Background()
	err := s.container.Terminate(ctx)
	require.NoError(s.T(), err)
	err = os.RemoveAll(s.cfgPath)
	require.NoError(s.T(), err)
}

func (s *DatabaseTestSuite) TestDatabaseQueries() {
	ctx := context.Background()
	const DefaultQueryLimit int32 = 857

	s.Run("initial_words_add", func() {
		conn, err := pgx.Connect(ctx, s.connStr)
		require.NoError(s.T(), err)
		defer conn.Close(ctx)

		q := New(conn)
		for _, w := range testWords(s.T()) {
			row, err := q.CreateWord(ctx, w)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), w, row.Value)
		}
	})

	s.Run("list_words", func() {
		conn, err := pgx.Connect(ctx, s.connStr)
		require.NoError(s.T(), err)
		defer conn.Close(ctx)

		query := New(conn)
		params := ListWordsParams{Limit: DefaultQueryLimit}
		rows, err := query.ListWords(ctx, params)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(testWords(s.T())), len(rows))
	})

	s.Run("create_word", func() {
		conn, err := pgx.Connect(ctx, s.connStr)
		require.NoError(s.T(), err)
		defer conn.Close(ctx)

		q := New(conn)
		row, err := q.CreateWord(ctx, "test1")
		require.NoError(s.T(), err)
		require.Equal(s.T(), "test1", row.Value)
	})

	s.Run("list_word_frequencies", func() {
		conn, err := pgx.Connect(ctx, s.connStr)
		require.NoError(s.T(), err)
		defer conn.Close(ctx)

		q := New(conn)
		params := ListWordFrequenciesParams{Limit: DefaultQueryLimit}
		rows, err := q.ListWordFrequencies(ctx, params)
		require.NoError(s.T(), err)

		for _, row := range rows {
			switch row.Value {
			case "leading":
				require.Equal(s.T(), int64(2), row.Total)
			case "development":
				require.Equal(s.T(), int64(10), row.Total)
			case "experience":
				require.Equal(s.T(), int64(28), row.Total)
			}
		}
	})

	s.Run("list_word_rankings", func() {
		conn, err := pgx.Connect(ctx, s.connStr)
		require.NoError(s.T(), err)
		defer conn.Close(ctx)

		q := New(conn)
		params := ListWordRankingsParams{Limit: DefaultQueryLimit}
		rows, err := q.ListWordRankings(ctx, params)
		require.NoError(s.T(), err)

		for _, row := range rows {
			switch row.Value {
			case "experience":
				require.Equal(s.T(), int64(1), row.Ranking)
			case "team":
				require.Equal(s.T(), int64(2), row.Ranking)
			case "software":
				require.Equal(s.T(), int64(3), row.Ranking)
			}
		}
	})

	s.Run("list_word_batches", func() {
		conn, err := pgx.Connect(ctx, s.connStr)
		require.NoError(s.T(), err)
		defer conn.Close(ctx)

		q := New(conn)
		params := ListWordBatchesParams{Limit: DefaultQueryLimit}
		batchRows, err := q.ListWordBatches(ctx, params)
		require.NoError(s.T(), err)

		for _, row := range batchRows {
			s.T().Logf("Got batch: %v", row)
		}
	})

	s.Run("list_words_by_batch_name_returns_batch_ids_and_related_words", func() {
		conn, err := pgx.Connect(ctx, s.connStr)
		require.NoError(s.T(), err)
		defer conn.Close(ctx)

		q := New(conn)
		rows, err := q.ListWordsByBatchName(ctx, "test_batch")
		require.NoError(s.T(), err)
		require.Empty(s.T(), rows)
	})
}

func (s *DatabaseTestSuite) TestCreatePhrasesBatchQuery() {
	ctx := context.Background()

	s.Run("creates_a_valid_phrases_batch_row", func() {
		conn, err := pgx.Connect(ctx, s.connStr)
		require.NoError(s.T(), err)
		defer conn.Close(ctx)

		row := conn.QueryRow(ctx, createPhrasesBatch, "test", loadTestPhrases(s.T()))
		var i CreatePhrasesBatchRow
		err = row.Scan(&i.ID, &i.BatchID)
		require.NoError(s.T(), err)
		s.T().Logf("Created phrase batch: %v", i.BatchID)
	})
}

// Helper functions remain the same
func loadTestPhrases(t *testing.T) []string {
	t.Helper()

	f, err := os.Open(filepath.Join("testdata", "phrases.txt"))
	require.NoError(t, err)
	defer f.Close()

	lines := make([]string, 0)
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		panic(err)
	}
	return lines
}

func testWords(t *testing.T) []string {
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

func applyMigrations(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	m, err := migrate.New("file://"+filepath.Join("testdata", "migrations"), url)
	if err != nil {
		return fmt.Errorf("new migrate: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up migrations: %w", err)
	}

	return nil
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
  name: piccrack
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

func (tc *testConfig) Remove() error {
	if err := os.RemoveAll(tc.f.Name()); err != nil {
		return fmt.Errorf("remove all: %w", err)
	}
	return nil
}
