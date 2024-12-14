package words

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/kndrad/piccrack/cmd/logger"
	"github.com/kndrad/piccrack/config"
	"github.com/kndrad/piccrack/internal/database"
	"github.com/kndrad/piccrack/pkg/retry"
	"github.com/spf13/cobra"
)

var Verbose bool

var rootCmd = &cobra.Command{
	Use:     "words",
	Short:   "Lists words from a database",
	Example: "piccrack words [OPTIONAL args: limit[int32]]",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		cfg, err := config.Load("config/development.yaml")
		if err != nil {
			l.Error("Loading onfig", "err", err.Error())

			return fmt.Errorf("load config: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

		q := database.New(conn)

		limit := math.MaxInt32
		if len(args) > 0 {
			limitArg, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("parse uint: %w", err)
			}
			if limitArg < math.MaxInt32 {
				limit = int(limitArg)
			}
		}
		params := database.ListWordsParams{Limit: int32(limit)}

		if len(args) > 0 {
			limit, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				l.Error("Failed to strconv", "err", err.Error())
			}
			params.Limit = int32(limit)
		}
		words, err := q.ListWords(ctx, params)
		if err != nil {
			l.Error("Connecting to database", "err", err.Error())

			return fmt.Errorf("database connection: %w", err)
		}
		l.Info("Listing words from a database", "len_words", len(words))
		for _, word := range words {
			fmt.Printf("%v\n", word)
		}

		return nil
	},
}

func RootCmd() *cobra.Command {
	return rootCmd
}
