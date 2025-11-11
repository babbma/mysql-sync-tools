package test

import (
	"testing"

	"github.com/db-sync-tools/internal/config"
	"gopkg.in/yaml.v3"
)

func TestConfigValidate_Success(t *testing.T) {
	yml := `
datasources:
  dev:
    url: "127.0.0.1:3306/dev_db"
    username: "root"
    password: "dev"
  prod:
    url: "192.168.1.100:3306/prod_db"
    username: "admin"
    password: "prod"
sync:
  source: "prod"
  target: "dev"
  batch_size: 200
  timeout: 3600
  truncate_before_sync: false
log:
  level: "info"
  console: true
`
	var cfg config.Config
	if err := yaml.Unmarshal([]byte(yml), &cfg); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected validate success, got error: %v", err)
	}

	if cfg.GetSourceConfig() == nil || cfg.GetTargetConfig() == nil {
		t.Fatalf("expected non-nil source/target config")
	}
	if cfg.GetSourceConfig().URL != "192.168.1.100:3306/prod_db" {
		t.Errorf("unexpected source url: %s", cfg.GetSourceConfig().URL)
	}
	if cfg.GetTargetConfig().URL != "127.0.0.1:3306/dev_db" {
		t.Errorf("unexpected target url: %s", cfg.GetTargetConfig().URL)
	}
}

func TestConfigValidate_MissingAliases(t *testing.T) {
	yml := `
datasources:
  only:
    url: "127.0.0.1:3306/only"
    username: "root"
    password: "pwd"
sync:
  source: "prod"
  target: "dev"
  batch_size: 100
`
	var cfg config.Config
	if err := yaml.Unmarshal([]byte(yml), &cfg); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validate error for missing aliases, got nil")
	}
}

func TestDatabaseConfig_GetDSN_WithDatabase(t *testing.T) {
	d := config.DatabaseConfig{
		URL:      "host.example.com:3306/mydb",
		Username: "u",
		Password: "p",
	}
	dsn := d.GetDSN()
	wantPrefix := "u:p@tcp(host.example.com:3306)/mydb?"
	if len(dsn) < len(wantPrefix) || dsn[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("unexpected dsn: %s", dsn)
	}
}

func TestDatabaseConfig_GetDSN_WithoutDatabase(t *testing.T) {
	d := config.DatabaseConfig{
		URL:      "host.example.com:3306",
		Username: "u",
		Password: "p",
	}
	dsn := d.GetDSN()
	wantPrefix := "u:p@tcp(host.example.com:3306)/?"
	if len(dsn) < len(wantPrefix) || dsn[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("unexpected dsn (no db): %s", dsn)
	}
}
