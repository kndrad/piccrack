package config_test

import (
	"os"
	"testing"

	"github.com/kndrad/piccrack/config"
	"github.com/stretchr/testify/require"
)

func TestLoadingConfig(t *testing.T) {
	// Create tmp yaml
	tmf, err := os.CreateTemp(t.TempDir(), "*.yaml")
	defer tmf.Close()
	defer os.RemoveAll(tmf.Name())

	// Write
	data := []byte(`app:
  environment: "testing"

http:
  host: "0.0.0.0"
  port: "8080"
  tls_enabled: false

database:
  user: testuser
  password: testpassword
  host: localhost
  port: 5433
  name: piccrack
  pool:
    max_conns: 25
    min_conns: 5
    max_conn_lifetime: 1h
    max_conn_idle_time: 30m
    connect_timeout: 10s
    dialer_keep_alive: 5s
`)
	if _, err := tmf.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// Load config
	cfg, err := config.Load(tmf.Name())
	require.NoError(t, err)
	// Checks
	require.Equal(t, "testing", cfg.App.Environment)

	require.Equal(t, "0.0.0.0", cfg.HTTP.Host)
	require.Equal(t, "8080", cfg.HTTP.Port)
	require.Equal(t, false, cfg.HTTP.TLSEnabled)

	require.Equal(t, "testuser", cfg.Database.User)
	require.Equal(t, "testpassword", cfg.Database.Password)
	require.Equal(t, "localhost", cfg.Database.Host)
	require.Equal(t, "5433", cfg.Database.Port)
	require.Equal(t, "piccrack", cfg.Database.Name)

	require.Equal(t, 25, cfg.Database.Pool.MaxConns)
	require.Equal(t, 5, cfg.Database.Pool.MinConns)
	require.Equal(t, "1h", cfg.Database.Pool.MaxConnLifetime)
	require.Equal(t, "30m", cfg.Database.Pool.MaxConnIdleTime)
	require.Equal(t, "10s", cfg.Database.Pool.ConnectTimeout)
	require.Equal(t, "5s", cfg.Database.Pool.DialerKeepAlive)
}
