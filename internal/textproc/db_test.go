package textproc_test

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
	"github.com/kndrad/wordcrack/internal/textproc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostgresDatabase(t *testing.T) {
	t.Parallel()

	tmpFile, err := os.CreateTemp("testdata", "*.env")
	require.NoError(t, err)
	content := `DB_USER=postgres
DB_PASSWORD=testpassword
DB_HOST=localhost
DB_PORT=5433
DB_NAME=itcrack`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write content, err: %v", err)
	}

	config, err := textproc.LoadDatabaseConfig(tmpFile.Name())
	require.NoError(t, err)
	if err := os.RemoveAll(tmpFile.Name()); err != nil {
		t.Fatalf("Failed to remove tmp file, err: %v", err)
	}

	ctx := context.Background()
	dbContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithUsername(config.User),
		postgres.WithPassword(config.Password),
		postgres.WithDatabase(config.DBName),
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
	if strings.HasSuffix(wd, "internal/textproc") {
		// Remove last element from wd
		cut := func(s, dir string) string {
			path, ok := strings.CutSuffix(s, dir)
			if ok {
				return path
			}
			t.Logf("Didnt cut, because '%s' was not found in %s", dir, s)

			return s
		}
		root = cut(wd, "internal/textproc")
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

	t.Run("insert_words", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := textproc.New(conn)
		for _, w := range words {
			row, err := q.InsertWord(ctx, w)
			require.NoError(t, err)
			assert.Equal(t, w, row.Value)
		}
	})

	t.Run("all_words", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := textproc.New(conn)
		params := textproc.AllWordsParams{Limit: DefaultQueryLimit}
		rows, err := q.AllWords(ctx, params)
		require.NoError(t, err)
		require.Equal(t, len(words), len(rows))
	})

	t.Run("add_word", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := textproc.New(conn)
		row, err := q.InsertWord(ctx, "test1")
		require.NoError(t, err)
		require.Equal(t, "test1", row.Value)
	})

	t.Run("words_frequency_count", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := textproc.New(conn)
		params := textproc.GetWordsFrequenciesParams{Limit: DefaultQueryLimit}
		rows, err := q.GetWordsFrequencies(ctx, params)
		require.NoError(t, err)

		for _, row := range rows {
			// Pick some random words
			switch row.Value {
			case "leading":
				require.Equal(t, int64(2), row.Frequency)
			case "development":
				require.Equal(t, int64(10), row.Frequency)
			case "experience":
				require.Equal(t, int64(28), row.Frequency)
			}
		}
	})

	t.Run("words_rank", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, connStr)
		require.NoError(t, err)
		defer conn.Close(ctx)

		q := textproc.New(conn)
		params := textproc.GetWordsRankParams{Limit: DefaultQueryLimit}
		rows, err := q.GetWordsRank(ctx, params)
		require.NoError(t, err)

		for _, row := range rows {
			// Pick some random words
			switch row.Value {
			case "experience":
				require.Equal(t, int64(1), row.Rank)
			case "team":
				require.Equal(t, int64(2), row.Rank)
			case "software":
				require.Equal(t, int64(3), row.Rank)
			}
		}
	})
}
