package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	App       AppConfig       `mapstructure:"app"`
	Seed      SeedConfig      `mapstructure:"seed"`
	Simulator SimulatorConfig `mapstructure:"simulator"`
}

type ServerConfig struct {
	Host         string `mapstructure:"host" validate:"required"`
	Port         string `mapstructure:"port" validate:"required"`
	ReadTimeout  int    `mapstructure:"read_timeout" validate:"gte=0"`
	WriteTimeout int    `mapstructure:"write_timeout" validate:"gte=0"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"name" validate:"required"`
}

type AppConfig struct {
	Environment string `mapstructure:"environment" validate:"required,oneof=local development production"`
	LogLevel    string `mapstructure:"log_level" validate:"required,oneof=debug info warn error fatal"`
}

type SeedConfig struct {
	Terminals int `mapstructure:"terminals" validate:"gte=2"`
}

type SimulatorConfig struct {
	APIURL           string  `mapstructure:"api_url" validate:"required,url"`
	IntervalSeconds  int     `mapstructure:"interval_seconds" validate:"gt=0"`
	StatusFlipChance float64 `mapstructure:"status_flip_chance" validate:"gte=0,lte=1"`
	MaxTransferBytes int64   `mapstructure:"max_transfer_bytes" validate:"gt=0"`
}

var cfg *Config

// setDefaults doubles as the key registry: AutomaticEnv only resolves keys viper
// already knows about, so a key absent here is not overridable by env var.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", 10)
	v.SetDefault("server.write_timeout", 10)

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "observv")

	v.SetDefault("app.environment", "local")
	v.SetDefault("app.log_level", "debug")

	v.SetDefault("seed.terminals", 10)

	v.SetDefault("simulator.api_url", "http://127.0.0.1:8080")
	v.SetDefault("simulator.interval_seconds", 5)
	v.SetDefault("simulator.status_flip_chance", 0.1)
	v.SetDefault("simulator.max_transfer_bytes", 100_000_000)
}

func Load() error {
	v := viper.New()

	setDefaults(v)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	c := &Config{}
	if err := v.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validator.New().Struct(c); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	cfg = c
	return nil
}

func Get() *Config {
	if cfg == nil {
		panic("config not loaded")
	}
	return cfg
}
