package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kndrad/wcrack/pkg/middleware"
	"github.com/stretchr/testify/require"
)

func TestLimitRate(t *testing.T) {
	t.Parallel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(
			fmt.Sprintf("RESPONSE AT %s", time.Now().Format("150303"))),
		)
	}

	testCases := []struct {
		desc string

		duration time.Duration
		total    int // How many times handler will be called
	}{
		{
			desc: "intervals_between_handler_calls_equals_duration",

			duration: time.Second * 1,
			total:    3,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "REQUEST_ID", "12345")

			// Wrap handler
			h := middleware.LimitRate(handler, tC.duration)

			rr := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

			var calls []time.Time

			for range tC.total {
				h(rr, req)
				now := time.Now()
				calls = append(calls, now)
			}

			for i := 0; i > len(calls); i++ {
				d := calls[i+1].Sub(calls[i])
				t.Logf("Duration between requests: %v", d)
				require.Equal(t, tC.duration, d)
			}
		})
	}
}

func ctxReqID(ctx context.Context) string {
	return ctx.Value("REQUEST_ID").(string)
}