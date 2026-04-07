package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := loadConfig("/nonexistent/path/config.json")
	if err != nil {
		t.Errorf("expected no error for nonexistent file, got: %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config for nonexistent file, got: %+v", cfg)
	}
}

func TestLoadConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	validConfig := `{
		"username": "testuser",
		"password": "testpass",
		"proxy": "http://proxy.example.com:8080",
		"verbose": true
	}`

	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}

	if cfg.Username != "testuser" {
		t.Errorf("expected username 'testuser', got: %s", cfg.Username)
	}
	if cfg.Password != "testpass" {
		t.Errorf("expected password 'testpass', got: %s", cfg.Password)
	}
	if cfg.Proxy != "http://proxy.example.com:8080" {
		t.Errorf("expected proxy 'http://proxy.example.com:8080', got: %s", cfg.Proxy)
	}
	if !cfg.Verbose {
		t.Error("expected verbose to be true")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	invalidJSON := `{ invalid json content }`

	if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if cfg != nil {
		t.Errorf("expected nil config for invalid JSON, got: %+v", cfg)
	}
}

func TestLoadConfig_InvalidProxyURL(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-proxy.json")

	invalidProxy := `{
		"username": "testuser",
		"proxy": "://invalid-scheme"
	}`

	if err := os.WriteFile(configPath, []byte(invalidProxy), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid proxy URL")
	}
	if cfg != nil {
		t.Errorf("expected nil config for invalid proxy, got: %+v", cfg)
	}
}

func TestLoadConfig_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.json")

	emptyConfig := `{}`

	if err := os.WriteFile(configPath, []byte(emptyConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Errorf("expected no error for empty config, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected empty config, got nil")
	}
}

func TestLoadConfig_CommentStripping(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "comments.json")

	configWithComments := `{
		# This is a comment
		"username": "testuser", // inline comment
		// Another comment
		"password": "testpass"
	}`

	if err := os.WriteFile(configPath, []byte(configWithComments), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Errorf("expected no error after stripping comments, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.Username != "testuser" {
		t.Errorf("expected username 'testuser', got: %s", cfg.Username)
	}
}

func TestValidateProxyURL(t *testing.T) {
	tests := []struct {
		url      string
		wantErr  bool
	}{
		{"http://proxy.example.com:8080", false},
		{"https://proxy.example.com:8080", false},
		{"socks5://proxy.example.com:1080", false},
		{"socks5h://proxy.example.com:1080", false},
		{"", false},
		{"ftp://proxy.example.com:8080", true},
		{"://invalid", true},
		{"http://", true},
		{"   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := validateProxyURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProxyURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_UsernameInvalidChars(t *testing.T) {
	cfg := &Config{
		Username: "user\nname",
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("expected error for username with invalid characters")
	}
}

func TestValidateConfig_DownloadParallelTooHigh(t *testing.T) {
	cfg := &Config{
		Download: &DownloadCfg{
			Parallel: 150,
		},
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("expected error for parallel > 100")
	}
}

func TestLoadConfig_DownloadSubconfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "download.json")

	downloadConfig := `{
		"download": {
			"parallel": 10,
			"out": "zones",
			"retries": 5,
			"exclude": "com,net"
		}
	}`

	if err := os.WriteFile(configPath, []byte(downloadConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if cfg == nil || cfg.Download == nil {
		t.Fatal("expected download config, got nil")
	}

	if cfg.Download.Parallel != 10 {
		t.Errorf("expected parallel 10, got %d", cfg.Download.Parallel)
	}
	if cfg.Download.OutDir != "zones" {
		t.Errorf("expected out 'zones', got %s", cfg.Download.OutDir)
	}
	if cfg.Download.Retries != 5 {
		t.Errorf("expected retries 5, got %d", cfg.Download.Retries)
	}
	if cfg.Download.Exclude != "com,net" {
		t.Errorf("expected exclude 'com,net', got %s", cfg.Download.Exclude)
	}
}

func TestLoadConfig_RequestSubconfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "request.json")

	requestConfig := `{
		"request": {
			"reason": "Research project",
			"request-all": true
		}
	}`

	if err := os.WriteFile(configPath, []byte(requestConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if cfg == nil || cfg.Request == nil {
		t.Fatal("expected request config, got nil")
	}

	if cfg.Request.Reason != "Research project" {
		t.Errorf("expected reason 'Research project', got %s", cfg.Request.Reason)
	}
	if !cfg.Request.RequestAll {
		t.Error("expected request-all to be true")
	}
}

func TestLoadConfig_StatusSubconfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "status.json")

	statusConfig := `{
		"status": {
			"zone": "com",
			"progress": true
		}
	}`

	if err := os.WriteFile(configPath, []byte(statusConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if cfg == nil || cfg.Status == nil {
		t.Fatal("expected status config, got nil")
	}

	if cfg.Status.Zone != "com" {
		t.Errorf("expected zone 'com', got %s", cfg.Status.Zone)
	}
	if !cfg.Status.Progress {
		t.Error("expected progress to be true")
	}
}

func TestConfigPaths(t *testing.T) {
	paths := configPaths()

	if len(paths) == 0 {
		t.Error("expected at least one config path")
	}

	if paths[0] == "" {
		t.Error("first config path should not be empty")
	}
}

func TestDefaultConfigDir(t *testing.T) {
	dir := defaultConfigDir()

	if dir == "" {
		t.Error("expected default config dir to be set")
	}

	expectedSuffix := ".czds"
	if len(dir) < len(expectedSuffix) || dir[len(dir)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("expected dir to end with %s, got %s", expectedSuffix, dir)
	}
}
