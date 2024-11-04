package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kndrad/itcrack/pkg/retry"
	"github.com/stretchr/testify/require"
)

// Implements retry.Pool interface.
type poolMock struct {
	calls     uint64
	shouldErr bool
	closed    bool
}

func (pmock *poolMock) Ping(ctx context.Context) error {
	pmock.calls++

	if pmock.shouldErr {
		return errors.New("ping failed")
	}

	return nil
}

func (pmock *poolMock) Close() {
	pmock.closed = true
}

func TestPingingDatabase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		retries       uint64
		shouldErr     bool
		expectedCalls uint64
	}{
		{
			name:          "success_on_first_try",
			retries:       retry.MaxRetries,
			shouldErr:     false,
			expectedCalls: 1,
		},
		{
			name:          "fails_on_no_more_retries",
			retries:       2,
			shouldErr:     true,
			expectedCalls: 3,
		},
		{
			name:          "passing_zero_retries_should_use_default",
			retries:       0,
			shouldErr:     true,
			expectedCalls: 4,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			pool := &poolMock{shouldErr: tC.shouldErr}

			ctx := context.Background()
			err := retry.PingDatabase(ctx, pool, tC.retries)

			if tC.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if pool.calls != tC.expectedCalls {
				t.Fatalf("expected %d calls, got %d", tC.expectedCalls, pool.calls)
			}
		})
	}
}

func TestPingDatabaseRespectsContextCancellation(t *testing.T) {
	t.Run("respects_context_cancellation", func(t *testing.T) {
		pool := &poolMock{shouldErr: true}
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context after small delay to check if it's respected
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		err := retry.PingDatabase(ctx, pool, retry.MaxRetries)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context cancellation error, got: %v", err)
		}
	})
}
