//go:build integration

package database

import (
	"bufio"
	"bytes"
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/kndrad/wcrack/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestDatabaseQueries(t *testing.T) {
	t.Parallel()

	tmpf, err := os.CreateTemp(t.TempDir(), "*.env")
	require.NoError(t, err)
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
	if _, err := tmpf.Write(content); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	config, err := config.Load(tmpf.Name())
	require.NoError(t, err)
	if err := os.RemoveAll(tmpf.Name()); err != nil {
		t.Fatalf("Failed to remove tmp file, err: %v", err)
	}

	ctx := context.Background()
	dbContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithUsername(config.Database.User),
		postgres.WithPassword(config.Database.Password),
		postgres.WithDatabase(config.Database.Name),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	require.NoError(t, err)
	defer func() {
		if err := testcontainers.TerminateContainer(dbContainer); err != nil {
			t.Fatalf("Failed to terminate container, err: %v", err)
		}
	}()

	// Change dir if wd points to 'tests' in project dir
	var root string
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Logf("Current wd: %s", wd)
	if strings.HasSuffix(wd, "internal/textproc/database") {
		// Remove last element from wd
		cut := func(s, dir string) string {
			path, ok := strings.CutSuffix(s, dir)
			if ok {
				return path
			}
			t.Logf("Didnt cut, because '%s' was not found in %s", dir, s)

			return s
		}
		root = cut(wd, "internal/textproc/database")
		t.Logf("Cut dir, got root: %s", root)
		if err := os.Chdir(root); err != nil {
			t.Fatalf("Failed to change dir, err: %v", err)
		}
	} else {
		root = wd
	}
	source := "file://" + root + "db/migrations"
	connStr, err := dbContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run migrations
	t.Logf("Running migrations, source: %s", source)
	m, err := migrate.New(source, connStr)
	require.NoError(t, err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Failed to run migrations, err %v", err)
	}

	// Fill the database with words
	data, err := os.ReadFile(filepath.Join("testdata", "words.txt"))
	require.NoError(t, err)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanWords)
	words := make([]string, 0)
	for scanner.Scan() {
		words = append(words, strings.Trim(scanner.Text(), " "))
	}
	t.Logf("Read %d words from a test words.txt file", len(words))

	if len(words) > math.MaxInt32 {
		t.Fatalf("words len %d exceeds max %d", len(words), math.MaxInt32)
	}
	const DefaultQueryLimit int32 = 857

	t.Run("", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := New(conn)
		for _, w := range words {
			row, err := q.CreateWord(ctx, w)
			require.NoError(t, err)
			assert.Equal(t, w, row.Value)
		}
	})

	t.Run("list_words", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
		require.NoError(t, err)
		defer conn.Close(ctx)

		query := New(conn)
		params := ListWordsParams{Limit: DefaultQueryLimit}
		rows, err := query.ListWords(ctx, params)
		require.NoError(t, err)
		require.Equal(t, len(words), len(rows))
	})

	t.Run("create_word", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := New(conn)
		row, err := q.CreateWord(ctx, "test1")
		require.NoError(t, err)
		require.Equal(t, "test1", row.Value)
	})

	t.Run("list_word_frequencies", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
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
		conn, err := pgx.Connect(ctx, connStr)
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
		conn, err := pgx.Connect(ctx, connStr)
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
}
