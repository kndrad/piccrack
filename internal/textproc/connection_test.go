package textproc_test

import (
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

	tmpf, err := os.CreateTemp("", "config*.env")
	require.NoError(t, err)
	defer os.Remove(tmpf.Name())

	content := `
	DB_USER=test_user
	DB_PASSWORD=test_password
	DB_HOST=localhost
	DB_PORT=5432
	DB_NAME=test_db
`

	if _, err := tmpf.WriteString(content); err != nil {
		t.Fatalf("failed to write string: %s, err: %v", content, err)
	}
	require.NoError(t, tmpf.Close())

	cfg, err := textproc.LoadDatabaseConfig(tmpf.Name())
	require.NoError(t, err)
	assert.Equal(t, "test_user", cfg.User)
	assert.Equal(t, "test_password", cfg.Password)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "5432", cfg.Port)
	assert.Equal(t, "test_db", cfg.DBName)
}
