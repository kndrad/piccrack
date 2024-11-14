package api

import (
	"context"
	"log/slog"
	"net/http"
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
		wantErr            bool
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

		srv    *httpServer
		signal os.Signal
	}{
		{
			desc:   "stops_after_interrupt_signal",
			srv:    NewHTTPServer(newTestCfg(t), newTestLogger(t)),
			signal: os.Interrupt,
		},
		{
			desc:   "stops_after_syscall_sigterm_signal",
			srv:    NewHTTPServer(newTestCfg(t), newTestLogger(t)),
			signal: syscall.SIGTERM,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx := context.Background()
			go tC.srv.Start(ctx)

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

func TestClientCheckHealth(t *testing.T) {
	testCases := []struct {
		desc string

		client             *httpClient
		expectedStatusCode int

		mustErr bool
	}{
		{
			desc: "status_ok",

			client:             NewHTTPClient(newTestCfg(t), newTestLogger(t)),
			expectedStatusCode: http.StatusOK,

			mustErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Init server
			go initHTTPServer(t)

			err := tC.client.CheckHealth(context.TODO())

			if tC.mustErr {
				require.Error(t, err)
			}
			require.NoError(t, err)

			sendKill(t)
		})
	}
}

func initHTTPServer(t *testing.T) error {
	t.Helper()

	srv := NewHTTPServer(newTestCfg(t), newTestLogger(t))

	return srv.Start(context.TODO())
}

func sendKill(t *testing.T) {
	t.Helper()

	// Send signal
	pid, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	signal := syscall.SIGTERM
	if err := pid.Signal(signal); err != nil {
		t.Fatalf("Sending signal %v failed: %v", signal, err)
		t.FailNow()
	}
}
