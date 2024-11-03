package textproc_test

import (
	"testing"

	"github.com/kndrad/itcrack/internal/textproc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConfigValidation(t *testing.T) {
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
