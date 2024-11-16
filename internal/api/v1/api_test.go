package v1

import (
	"context"
	"log/slog"
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

func newTestCfg(t *testing.T) *Config {
	t.Helper()

	cfg := &Config{
		Host:       "localhost",
		Port:       "8080",
		TLSEnabled: false,
	}

	return cfg
}

func newTestLogger(t *testing.T) *slog.Logger {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	return logger
}

func TestHTTPServerStart(t *testing.T) {
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
			srv, err := NewHTTPServer(newTestCfg(t), newTestLogger(t))
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

func TestCheckHealthHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		expectedStatusCode int
		mustErr            bool
	}{
		{
			desc: "status_ok",

			expectedStatusCode: http.StatusOK,
			mustErr:            false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Init server
			ts := httptest.NewServer(http.HandlerFunc(checkHealthHandler))
			defer ts.Close()

			resp, err := ts.Client().Get(ts.URL)
			resp.Body.Close()

			if tC.mustErr {
				require.Error(t, err)
			}

			require.NoError(t, err)
			require.Equal(t, tC.expectedStatusCode, resp.StatusCode)
		})
	}
}

func TestClientCheckHealth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		cfg                *Config
		expectedStatusCode int
		mustErr            bool
	}{
		{
			desc: "status_ok",

			expectedStatusCode: http.StatusOK,
			mustErr:            false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c := NewClient(newTestCfg(t), newTestLogger(t))
			defer c.Close()

			err := c.CheckHealth(context.Background())

			if tC.mustErr {
				require.Error(t, err)
			}
			require.NoError(t, err)
		})
	}
}

func TestRoutes(t *testing.T) {
	base := newTestCfg(t).BaseURL()

	r := new(routes)
	r.Add(base, "test")

	require.NotNil(t, r)

	testRoute, err := r.Get("test")
	require.NoError(t, err)
	require.NotEmpty(t, testRoute)
	require.Contains(t, testRoute, "test")
}
