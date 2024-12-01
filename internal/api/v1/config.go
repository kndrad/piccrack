package v1

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host       string `mapstructure:"HTTP_HOST"`
	Port       string `mapstructure:"HTTP_PORT"`
	TLSEnabled bool   `mapstructure:"TLS_ENABLED"`
}

func LoadConfig(path string) (Config, error) {
	path = filepath.Clean(path)
	viper.SetConfigFile(path)

	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open err: %w", err)
	}
	defer f.Close()

	if err := viper.ReadConfig(f); err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	cfg := Config{
		Host:       viper.GetString("HTTP_HOST"),
		Port:       viper.GetString("HTTP_PORT"),
		TLSEnabled: viper.GetBool("TLS_ENABLED"),
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func (cfg Config) Addr() string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}

func (cfg Config) BaseURL() string {
	sep := ":" + string(filepath.Separator) + string(filepath.Separator)

	var urlPrefix string
	switch cfg.TLSEnabled {
	case true:
		urlPrefix = "https"
	case false:
		urlPrefix = "http"
	}

	return urlPrefix + sep + cfg.Addr()
}
