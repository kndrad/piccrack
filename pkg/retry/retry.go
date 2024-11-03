package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type Pool interface {
	Ping(ctx context.Context) error
	Close()
}

const MaxRetries uint64 = 3

func PingDatabase(ctx context.Context, pool Pool, retries uint64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	op := func() error {
		if err := pool.Ping(ctx); err != nil {
			pool.Close()

			return fmt.Errorf("pinging db: %w", err)
		}

		return nil
	}
	if retries == 0 {
		retries = MaxRetries
	}
	b := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), retries)
	if err := backoff.Retry(op, b); err != nil {
		return fmt.Errorf("retrying operation: %w", err)
	}

	return nil
}
