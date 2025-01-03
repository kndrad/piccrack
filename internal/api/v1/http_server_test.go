package v1

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kndrad/piccrack/config"
	"github.com/stretchr/testify/require"
)

func mockConfig() config.API {
	cfg := config.API{
		Host:       "localhost",
		Port:       "8080",
		TLSEnabled: false,
	}

	return cfg
}

func TestServerStart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		signal os.Signal
	}{
		{
			desc: "stops_after_interrupt_signal",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Context with cancel
			ctx, cancel := context.WithCancel(context.Background())

			// Create server instance
			srv, err := New(
				mockConfig(),
				&service{
					q:      NewQueriesMock(NewWordsMock()...),
					logger: testLogger(),
				},
				testLogger(),
			)
			require.NoError(t, err)

			// Done channel to wait for server shutdown
			started := make(chan struct{})

			// Start server in goroutine
			go func() {
				err := srv.Start(ctx)
				if err != nil && !errors.Is(err, context.Canceled) {
					t.Errorf("Server start err: %v", err)
				}
				started <- struct{}{}
				close(started)
			}()
			cancel()
			<-started
		})
	}
}

func TestWriteJSONErr(t *testing.T) {
	t.Parallel()

	errOpFailed := errors.New("operation: failed")

	testCases := []struct {
		desc string

		msg         string
		err         error
		code        int
		contentType string
	}{
		{
			desc: "writes_msg_err_code_with_header",

			msg:         "test message",
			err:         errOpFailed,
			code:        http.StatusBadRequest,
			contentType: "application/json",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			rr := httptest.NewRecorder()

			// Write
			respondJSON(rr, tC.msg, tC.err, tC.code)

			// Get result
			resp := rr.Result()
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Check Content-Type header
			contentType := resp.Header.Get("Content-Type")
			require.Equal(t, tC.contentType, contentType)

			t.Logf("data: %v", string(data))

			// Check if data contains desired message and error
			require.Contains(t, string(data), tC.msg)
			require.Contains(t, string(data), tC.err.Error())

			// Check code
			require.Equal(t, tC.code, resp.StatusCode)
		})
	}
}
