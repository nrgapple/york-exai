package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"yorkexai/internal/store"
)

const configVersion = 1

type IntegrationConfig struct {
	Status string `json:"status,omitempty"`
	Mode   string `json:"mode,omitempty"`
	Notes  string `json:"notes,omitempty"`
}

type Config struct {
	Version      int                          `json:"version"`
	Storage      map[string]string            `json:"storage,omitempty"`
	Integrations map[string]IntegrationConfig `json:"integrations,omitempty"`
}

type Paths struct {
	Home         string
	ConfigPath   string
	StateDir     string
	DBPath       string
	ArtifactsDir string
	AudioDir     string
	PhotosDir    string
	DocumentsDir string
	ExportsDir   string
	BackupsDir   string
	TempDir      string
}

type Runtime struct {
	Paths  Paths
	Config Config
	Store  *store.Store
}

func ResolveHome(homeOverride string) (string, error) {
	if homeOverride != "" {
		return filepath.Clean(homeOverride), nil
	}

	if envHome := os.Getenv("YORK_HOME"); envHome != "" {
		return filepath.Clean(envHome), nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(base, "YorkExAI"), nil
}

func BuildPaths(home string) Paths {
	return Paths{
		Home:         home,
		ConfigPath:   filepath.Join(home, "config.json"),
		StateDir:     filepath.Join(home, "state"),
		DBPath:       filepath.Join(home, "state", "york.db"),
		ArtifactsDir: filepath.Join(home, "artifacts"),
		AudioDir:     filepath.Join(home, "artifacts", "audio"),
		PhotosDir:    filepath.Join(home, "artifacts", "photos"),
		DocumentsDir: filepath.Join(home, "artifacts", "documents"),
		ExportsDir:   filepath.Join(home, "artifacts", "exports"),
		BackupsDir:   filepath.Join(home, "backups"),
		TempDir:      filepath.Join(home, "tmp"),
	}
}

func MinimalConfig() Config {
	return Config{
		Version:      configVersion,
		Storage:      map[string]string{},
		Integrations: map[string]IntegrationConfig{},
	}
}

func LoadOrCreateRuntime(ctx context.Context, homeOverride string, autoInit bool) (*Runtime, error) {
	home, err := ResolveHome(homeOverride)
	if err != nil {
		return nil, err
	}

	paths := BuildPaths(home)
	if autoInit {
		if err := ensureDirs(paths); err != nil {
			return nil, err
		}
	}

	cfg, err := loadConfig(paths.ConfigPath, autoInit)
	if err != nil {
		return nil, err
	}

	st, err := store.Open(ctx, paths.DBPath, autoInit)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		Paths:  paths,
		Config: cfg,
		Store:  st,
	}, nil
}

func ensureDirs(paths Paths) error {
	for _, dir := range []string{
		paths.Home,
		paths.StateDir,
		paths.ArtifactsDir,
		paths.AudioDir,
		paths.PhotosDir,
		paths.DocumentsDir,
		paths.ExportsDir,
		paths.BackupsDir,
		paths.TempDir,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	return nil
}

func loadConfig(path string, autoCreate bool) (Config, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
		if cfg.Storage == nil {
			cfg.Storage = map[string]string{}
		}
		if cfg.Integrations == nil {
			cfg.Integrations = map[string]IntegrationConfig{}
		}
		return cfg, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	if !autoCreate {
		return Config{}, fmt.Errorf("config missing at %s", path)
	}

	cfg := MinimalConfig()
	if err := writeConfig(path, cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func writeConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
