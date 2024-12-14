package words

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kndrad/piccrack/cmd/logger"
	"github.com/kndrad/piccrack/config"
	"github.com/kndrad/piccrack/internal/database"
	"github.com/kndrad/piccrack/pkg/retry"
	"github.com/kndrad/piccrack/pkg/textproc"
	"github.com/spf13/cobra"
)

var addManyCmd = &cobra.Command{
	Use:     "many",
	Short:   "Adds many words to a database.",
	Example: "piccrack words add many [FILE PATH <name>.txt | <name>.json]",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		cfg, err := config.Load("config/development.yaml")
		if err != nil {
			l.Error("Loading database config", "err", err.Error())

			return fmt.Errorf("config load: %w", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		pool, err := database.Pool(ctx, cfg.Database)
		if err != nil {
			l.Error("Loading database pool", "err", err.Error())

			return fmt.Errorf("database pool: %w", err)
		}
		defer pool.Close()

		if err := retry.Ping(ctx, pool, retry.MaxRetries); err != nil {
			l.Error("Pinging database", "err", err.Error())

			return fmt.Errorf("database ping: %w", err)
		}

		conn, err := database.Connect(ctx, pool)
		if err != nil {
			l.Error("Connecting to database", "err", err.Error())

			return fmt.Errorf("database connection: %w", err)
		}
		defer conn.Close(ctx)

		// Read words from a json file
		analysis := new(textproc.TextAnalysis)
		path := filepath.Clean(args[0])

		switch filepath.Ext(path) {
		case ".json":
			data, err := os.ReadFile(args[0])
			if err != nil {
				l.Error("Failed to read file",
					slog.String("path", path),
				)

				return fmt.Errorf("read file: %w", err)
			}
			if err := json.Unmarshal(data, &analysis); err != nil {
				l.Error("Failed to unmarshal json into analysis")

				return fmt.Errorf("unmarshal json: %w", err)
			}
		case ".txt":
			data, err := os.ReadFile(args[0])
			if err != nil {
				l.Error("Failed to read file",
					slog.String("path", path),
				)

				return fmt.Errorf("read file: %w", err)
			}
			scanner := bufio.NewScanner(bytes.NewReader(data))
			scanner.Split(bufio.ScanWords)

			for scanner.Scan() {
				word := strings.Trim(scanner.Text(), " ")
				analysis.IncWordCount(word)
			}

			if err := scanner.Err(); err != nil {
				l.Error("Scanner returned an error", "err", err.Error())

				return fmt.Errorf("scanner err: %w", err)
			}
		}
		if Verbose {
			printWords(analysis)
		}

		// Query db to insert each word
		q := database.New(conn)
		for word := range analysis.WordFrequency {
			row, err := q.CreateWord(ctx, word)
			if err != nil {
				l.Error("Failed to insert word",
					slog.String("word", word),
				)

				return fmt.Errorf("word insert: %w", err)
			}
			l.Info("Inserted row to a database",
				slog.Int64("id", row.ID),
				slog.String("word", row.Value),
			)
		}

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	addCmd.AddCommand(addManyCmd)
}

func printWords(analysis *textproc.TextAnalysis) {
	for word, frequency := range analysis.WordFrequency {
		fmt.Printf("WORD: %s, FREQUENCY: %s\n", word, strconv.Itoa(frequency))
	}
}
