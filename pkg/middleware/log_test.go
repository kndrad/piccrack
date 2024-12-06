package middleware

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogTime(t *testing.T) {
	t.Parallel()

	w := new(bytes.Buffer)

	l := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{}))
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("Sent %s\n", time.Now())))
	}

	ctx := context.WithValue(context.Background(), "duration", nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	wrapped := LogTime(handler, l)
	wrapped(rr, req)

	// Was the message recorded?
	data := w.String()

	require.Contains(t, data, "Finished duration")
}
