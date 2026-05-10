package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultAwarenessCycleStartDate(t *testing.T) {
	var c Config
	if c.AwarenessCycle.StartDate != "" {
		t.Fatalf("zero-value config should not set awareness cycle start date, got %q", c.AwarenessCycle.StartDate)
	}
}

func TestLoadSupportsLegacyNumericCycle(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "zero-api.yaml")
	content := "Name: zero-api\nHost: 0.0.0.0\nPort: 8888\nMysql:\n  DataSource: \"\"\nCycle: 200\n"
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var cfg Config
	if err := Load(configPath, &cfg); err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Cycle.TotalPoints != 200 {
		t.Fatalf("expected cycle total points 200, got %d", cfg.Cycle.TotalPoints)
	}
}

func TestLoadSupportsStructuredCycle(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "zero-api.yaml")
	content := "Name: zero-api\nHost: 0.0.0.0\nPort: 8888\nMysql:\n  DataSource: \"\"\nCycle:\n  TotalPoints: 240\n"
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var cfg Config
	if err := Load(configPath, &cfg); err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Cycle.TotalPoints != 240 {
		t.Fatalf("expected cycle total points 240, got %d", cfg.Cycle.TotalPoints)
	}
}
