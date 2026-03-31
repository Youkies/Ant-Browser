package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBackfillsLegacyConfig(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	legacyConfig := `
logging: {}
browser: {}
`
	if err := os.WriteFile(configPath, []byte(legacyConfig), 0o644); err != nil {
		t.Fatalf("鍐欏叆娴嬭瘯閰嶇疆澶辫触: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("鍔犺浇閰嶇疆澶辫触: %v", err)
	}

	if cfg.Database.Type != "sqlite" {
		t.Fatalf("Database.Type 鏈ˉ榻? got=%q", cfg.Database.Type)
	}
	if cfg.Database.SQLite.Path != "data/app.db" {
		t.Fatalf("Database.SQLite.Path 鏈ˉ榻? got=%q", cfg.Database.SQLite.Path)
	}
	if cfg.App.Name != "Ant Browser" {
		t.Fatalf("App.Name 鏈ˉ榻? got=%q", cfg.App.Name)
	}
	if cfg.Runtime.MaxMemoryMB != 0 || cfg.Runtime.GCPercent != 100 {
		t.Fatalf("Runtime 鏈ˉ榻? got=%+v", cfg.Runtime)
	}
	if cfg.Logging.Level != "info" || cfg.Logging.FilePath != "data/logs/app.log" {
		t.Fatalf("Logging 鍩虹瀛楁鏈ˉ榻? got=%+v", cfg.Logging)
	}
	if !cfg.Logging.Interceptor.Enabled || !cfg.Logging.Interceptor.LogParameters || !cfg.Logging.Interceptor.LogResults {
		t.Fatalf("Interceptor 榛樿鍊兼湭琛ラ綈: got=%+v", cfg.Logging.Interceptor)
	}
	if len(cfg.Logging.Interceptor.SensitiveFields) == 0 {
		t.Fatalf("Interceptor.SensitiveFields 未补齐")
	}
	if cfg.Browser.UserDataRoot != "data" {
		t.Fatalf("Browser.UserDataRoot 鏈ˉ榻? got=%q", cfg.Browser.UserDataRoot)
	}
	if len(cfg.Browser.DefaultFingerprintArgs) == 0 || len(cfg.Browser.DefaultLaunchArgs) == 0 {
		t.Fatalf("Browser 默认启动参数未补齐")
	}
	if cfg.Browser.Cores == nil || cfg.Browser.Proxies == nil || cfg.Browser.Profiles == nil {
		t.Fatalf("Browser 鍒楄〃瀛楁搴斿垵濮嬪寲涓虹┖鍒囩墖")
	}
	if cfg.LaunchServer.Port != DefaultLaunchServerPort {
		t.Fatalf("LaunchServer.Port 鏈ˉ榻? got=%d", cfg.LaunchServer.Port)
	}
	if cfg.LaunchServer.Auth.Enabled {
		t.Fatalf("LaunchServer.Auth.Enabled 榛樿搴斾负 false: got=%v", cfg.LaunchServer.Auth.Enabled)
	}
	if cfg.LaunchServer.Auth.APIKey != "" {
		t.Fatalf("LaunchServer.Auth.APIKey 榛樿搴斾负绌? got=%q", cfg.LaunchServer.Auth.APIKey)
	}
	if cfg.LaunchServer.Auth.Header != DefaultLaunchServerAPIKeyHeader {
		t.Fatalf("LaunchServer.Auth.Header 鏈ˉ榻? got=%q", cfg.LaunchServer.Auth.Header)
	}
}

func TestLoadPreservesExplicitConfig(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	customConfig := `
database:
  type: sqlite
  sqlite:
    path: custom/app.db
app:
  name: Custom App
  window:
    width: 1400
    height: 800
    min_width: 900
    min_height: 600
runtime:
  max_memory_mb: 2048
  gc_percent: 80
logging:
  level: debug
  file_enabled: true
  file_path: custom.log
  format: json
  buffer_size: 8
  async_queue_size: 2000
  flush_interval_ms: 500
  rotation:
    enabled: true
    max_size_mb: 10
    max_age: 3
    max_backups: 2
    time_interval: hourly
  interceptor:
    enabled: false
    log_parameters: false
    log_results: false
    sensitive_fields: []
browser:
  user_data_root: custom_data
  default_fingerprint_args:
    - --fingerprint-brand=Edge
  default_launch_args:
    - --start-maximized
  default_proxy: direct://
  default_bookmarks: []
  cores: []
  proxies: []
  profiles: []
launch_server:
  port: 30000
  auth:
    enabled: true
    api_key: secret-key
    header: X-Custom-Ant-Key
`
	if err := os.WriteFile(configPath, []byte(customConfig), 0o644); err != nil {
		t.Fatalf("鍐欏叆娴嬭瘯閰嶇疆澶辫触: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("鍔犺浇閰嶇疆澶辫触: %v", err)
	}

	if cfg.App.Name != "Custom App" {
		t.Fatalf("App 鏄惧紡閰嶇疆琚鐩? got=%+v", cfg.App)
	}
	if cfg.Database.SQLite.Path != "custom/app.db" {
		t.Fatalf("Database.SQLite.Path 鏄惧紡閰嶇疆琚鐩? got=%q", cfg.Database.SQLite.Path)
	}
	if cfg.Runtime.MaxMemoryMB != 2048 || cfg.Runtime.GCPercent != 80 {
		t.Fatalf("Runtime 鏄惧紡閰嶇疆琚鐩? got=%+v", cfg.Runtime)
	}
	if cfg.Logging.Level != "debug" || cfg.Logging.Format != "json" || !cfg.Logging.FileEnabled {
		t.Fatalf("Logging 鏄惧紡閰嶇疆琚鐩? got=%+v", cfg.Logging)
	}
	if cfg.Logging.Interceptor.Enabled {
		t.Fatalf("Interceptor.Enabled 显式 false 被覆盖")
	}
	if len(cfg.Browser.DefaultFingerprintArgs) != 1 || cfg.Browser.DefaultFingerprintArgs[0] != "--fingerprint-brand=Edge" {
		t.Fatalf("Browser.DefaultFingerprintArgs 鏄惧紡閰嶇疆琚鐩? got=%v", cfg.Browser.DefaultFingerprintArgs)
	}
	if cfg.Browser.UserDataRoot != "custom_data" || cfg.Browser.DefaultProxy != "direct://" {
		t.Fatalf("Browser 鏄惧紡閰嶇疆琚鐩? got=%+v", cfg.Browser)
	}
	if cfg.LaunchServer.Port != 30000 {
		t.Fatalf("LaunchServer.Port 鏄惧紡閰嶇疆琚鐩? got=%d", cfg.LaunchServer.Port)
	}
	if !cfg.LaunchServer.Auth.Enabled {
		t.Fatalf("LaunchServer.Auth.Enabled 显式配置被覆盖")
	}
	if cfg.LaunchServer.Auth.APIKey != "secret-key" {
		t.Fatalf("LaunchServer.Auth.APIKey 鏄惧紡閰嶇疆琚鐩? got=%q", cfg.LaunchServer.Auth.APIKey)
	}
	if cfg.LaunchServer.Auth.Header != "X-Custom-Ant-Key" {
		t.Fatalf("LaunchServer.Auth.Header 鏄惧紡閰嶇疆琚鐩? got=%q", cfg.LaunchServer.Auth.Header)
	}
}

func TestLoadMigratesLegacyRootLogPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	legacyConfig := `
logging:
  file_path: logs/app.log
`
	if err := os.WriteFile(configPath, []byte(legacyConfig), 0o644); err != nil {
		t.Fatalf("鍐欏叆娴嬭瘯閰嶇疆澶辫触: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("鍔犺浇閰嶇疆澶辫触: %v", err)
	}

	if cfg.Logging.FilePath != "data/logs/app.log" {
		t.Fatalf("legacy 鏍圭洰褰曟棩蹇楄矾寰勬湭杩佺Щ: got=%q", cfg.Logging.FilePath)
	}
}
