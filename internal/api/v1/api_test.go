package v1

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)

	testCases := []struct {
		desc string

		fileName           string
		data               []byte
		expectedHost       string
		expectedPort       string
		expectedTLSEnabled bool
		expectedAddr       string
		expectedURLPrefix  string
		expectedBaseURL    string

		wantErr bool
	}{
		{
			desc: "reads_set_variables_in_env_file",

			fileName: "*.env",
			data: []byte(`HTTP_HOST="localhost"
HTTP_PORT="8080"
TLS_ENABLED=false`),
			expectedHost:       "localhost",
			expectedPort:       "8080",
			expectedTLSEnabled: false,
			expectedAddr:       "localhost:8080",
			expectedBaseURL:    "http://localhost:8080",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(wd, tC.fileName)
			require.NoError(t, err)
			if _, err := tmpFile.Write(tC.data); err != nil {
				t.Fatalf("Failed to write %v to %v", tC.data, tC.fileName)
			}
			cfg, err := LoadConfig(tmpFile.Name())
			if tC.wantErr {
				require.Error(t, err)
				require.Nil(t, cfg)
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)
			require.EqualValues(t, tC.expectedHost, cfg.Host)
			require.EqualValues(t, tC.expectedPort, cfg.Port)
			require.EqualValues(t, tC.expectedTLSEnabled, cfg.TLSEnabled)
			require.EqualValues(t, tC.expectedAddr, cfg.Addr())
			require.EqualValues(t, tC.expectedBaseURL, cfg.BaseURL())

			// Close
			if err := tmpFile.Close(); err != nil {
				t.Fatalf("Failed to close tmpFile: %v, err: %v", tmpFile.Name(), err)
			}
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Fatalf("Failed to remove tmpFile: %v, err: %v", tmpFile.Name(), err)
			}
		})
	}
}

func newTestCfg(t *testing.T) *ServerConfig {
	t.Helper()

	cfg := &ServerConfig{
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
			desc:   "stops_after_interrupt_signal",
			signal: os.Interrupt,
		},
		{
			desc:   "stops_after_syscall_sigterm_signal",
			signal: syscall.SIGTERM,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			srv, err := NewServer(
				newTestCfg(t),
				&WordService{q: mockWordQueries(wordsMock()...), logger: getTestLogger()},
				getTestLogger(),
			)
			require.NoError(t, err)

			ctx := context.Background()
			go srv.Start(ctx)

			<-time.After(1 * time.Second) // Wait to send signal

			// Send signal
			pid, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			if err := pid.Signal(tC.signal); err != nil {
				t.Fatalf("Sending signal %v failed: %v", tC.signal, err)
			}
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
			WriteJSONErr(rr, tC.msg, tC.err, tC.code)

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
