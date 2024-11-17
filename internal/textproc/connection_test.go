package textproc_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/kndrad/itcrack/internal/textproc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConfigValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		cfg     textproc.DatabaseConfig
		mustErr bool
	}{
		{
			desc: "valid config",
			cfg: textproc.DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "secret",
				DBName:   "testdb",
			},
		},
		{
			desc: "invalid config (one field empty)",
			cfg: textproc.DatabaseConfig{
				Host:   "localhost",
				Port:   "5432",
				User:   "postgres",
				DBName: "testdb",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := textproc.ValidateDatabaseConfig(tC.cfg)

			if tC.mustErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, textproc.ErrInvalidConfig)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadingDatabaseConfig(t *testing.T) {
	t.Parallel()

	tmpf, err := os.CreateTemp("testdata", "*.env")
	require.NoError(t, err)
	defer os.Remove(tmpf.Name())

	var (
		user   = "test_user"
		pswd   = "test_password"
		host   = "test_host"
		port   = "5489"
		dbName = "test_db"
	)

	content := fmt.Sprintf(`DB_USER=%s
DB_PASSWORD=%s
DB_HOST=%s
DB_PORT=%s
DB_NAME=%s
`, user, pswd, host, port, dbName)
	if _, err := tmpf.WriteString(content); err != nil {
		t.Fatalf("failed to write string: %s, err: %v", content, err)
	}
	require.NoError(t, tmpf.Close())

	cfg, err := textproc.LoadDatabaseConfig(tmpf.Name())
	require.NoError(t, err)
	require.Equal(t, user, cfg.User)
	require.Equal(t, pswd, cfg.Password)
	require.Equal(t, host, cfg.Host)
	require.Equal(t, port, cfg.Port)
	require.Equal(t, dbName, cfg.DBName)
}
