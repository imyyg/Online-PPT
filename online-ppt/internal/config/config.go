package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "configs/app.yaml"

// Config aggregates runtime settings loaded from YAML files or environment overrides.
type Config struct {
	Server   ServerConfig
	Security SecurityConfig
	Storage  StorageConfig
	Redis    RedisConfig
	SMTP     SMTPConfig
	Paths    PathConfig
}

// ServerConfig wraps HTTP server settings.
type ServerConfig struct {
	Addr string
}

// SecurityConfig covers JWT parameters and secrets.
type SecurityConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// StorageConfig stores database connectivity hints.
type StorageConfig struct {
	Driver string
	DSN    string
}

// PathConfig keeps filesystem root references.
type PathConfig struct {
	PresentationsRoot string
}

// RedisConfig stores Redis connectivity settings.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"`
}

// SMTPConfig stores email server settings.
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	FromName string `yaml:"fromName"`
	UseTLS   bool   `yaml:"useTLS"`
}

type securityRaw struct {
	JWTSecret       string `yaml:"jwtSecret"`
	AccessTokenTTL  string `yaml:"accessTokenTTL"`
	RefreshTokenTTL string `yaml:"refreshTokenTTL"`
}

type rawConfig struct {
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
	Security securityRaw `yaml:"security"`
	Storage  struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"storage"`
	Paths struct {
		PresentationsRoot string `yaml:"presentationsRoot"`
	} `yaml:"paths"`
}

// Load reads configuration from disk using APP_CONFIG_PATH override or default path.
func Load() (*Config, error) {
	path := os.Getenv("APP_CONFIG_PATH")
	if path == "" {
		path = defaultConfigPath
	}
	return LoadFromFile(path)
}

// LoadFromFile deserializes the YAML configuration located at path.
func LoadFromFile(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg := Config{
		Server: ServerConfig{Addr: raw.Server.Addr},
		Storage: StorageConfig{
			Driver: raw.Storage.Driver,
			DSN:    raw.Storage.DSN,
		},
		Paths: PathConfig{PresentationsRoot: raw.Paths.PresentationsRoot},
	}

	if cfg.Server.Addr == "" {
		return nil, errors.New("server.addr is required")
	}
	if cfg.Storage.Driver == "" {
		return nil, errors.New("storage.driver is required")
	}
	if cfg.Storage.DSN == "" {
		return nil, errors.New("storage.dsn is required")
	}
	if cfg.Paths.PresentationsRoot == "" {
		return nil, errors.New("paths.presentationsRoot is required")
	}

	if cfg.Security, err = parseSecurity(raw.Security); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func parseSecurity(sec securityRaw) (SecurityConfig, error) {
	if sec.JWTSecret == "" {
		return SecurityConfig{}, errors.New("security.jwtSecret is required")
	}

	accessTTL, err := time.ParseDuration(sec.AccessTokenTTL)
	if err != nil {
		return SecurityConfig{}, fmt.Errorf("parse security.accessTokenTTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(sec.RefreshTokenTTL)
	if err != nil {
		return SecurityConfig{}, fmt.Errorf("parse security.refreshTokenTTL: %w", err)
	}

	return SecurityConfig{
		JWTSecret:       sec.JWTSecret,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	}, nil
}
