package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configEnvVar    = "CZDS_CONFIG"
	configFileName   = ".czds.json"
	configDirName    = ".czds"
)

type Config struct {
	Username string       `json:"username,omitempty"`
	Password string       `json:"password,omitempty"`
	Proxy    string       `json:"proxy,omitempty"`
	Verbose  bool         `json:"verbose,omitempty"`
	Download *DownloadCfg `json:"download,omitempty"`
	Request  *RequestCfg  `json:"request,omitempty"`
	Status   *StatusCfg   `json:"status,omitempty"`
}

type DownloadCfg struct {
	Parallel   uint   `json:"parallel,omitempty"`
	OutDir     string `json:"out,omitempty"`
	URLName    bool   `json:"urlname,omitempty"`
	Force      bool   `json:"force,omitempty"`
	Redownload bool   `json:"redownload,omitempty"`
	Exclude    string `json:"exclude,omitempty"`
	Include    string `json:"include,omitempty"`
	Retries    uint   `json:"retries,omitempty"`
	Quiet      bool   `json:"quiet,omitempty"`
	Progress   bool   `json:"progress,omitempty"`
	DateDir    bool   `json:"datedir,omitempty"`
}

type RequestCfg struct {
	Reason     string `json:"reason,omitempty"`
	RequestTLDs string `json:"request,omitempty"`
	RequestAll  bool   `json:"request-all,omitempty"`
	Status     bool   `json:"status,omitempty"`
	ExtendTLDs string `json:"extend,omitempty"`
	ExtendAll   bool   `json:"extend-all,omitempty"`
	Exclude    string `json:"exclude,omitempty"`
	CancelTLDs string `json:"cancel,omitempty"`
}

type StatusCfg struct {
	ID       string `json:"id,omitempty"`
	Zone     string `json:"zone,omitempty"`
	Report   string `json:"report,omitempty"`
	Progress bool   `json:"progress,omitempty"`
}

func defaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, configDirName)
}

func configPaths() []string {
	var paths []string

	if envPath := os.Getenv(configEnvVar); envPath != "" {
		paths = append(paths, envPath)
	}

	if configDir := defaultConfigDir(); configDir != "" {
		paths = append(paths, filepath.Join(configDir, configFileName))
		paths = append(paths, filepath.Join(configDir, "config"))
	}

	paths = append(paths, configFileName)
	paths = append(paths, ".czds")

	return paths
}

func loadConfig(explicitPath string) (*Config, error) {
	var configPath string

	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); os.IsNotExist(err) {
			return nil, nil
		}
		configPath = explicitPath
	} else {
		for _, path := range configPaths() {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	if configPath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	data = stripComments(data)

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration in %s: %w", configPath, err)
	}

	return &cfg, nil
}

func stripComments(data []byte) []byte {
	var result []byte
	inString := false

	for i := 0; i < len(data); i++ {
		if data[i] == '"' && (i == 0 || data[i-1] != '\\') {
			inString = !inString
		}

		if !inString && i+1 < len(data) {
			if (data[i] == '/' && data[i+1] == '/') || data[i] == '#' {
				for i < len(data) && data[i] != '\n' {
					i++
				}
				continue
			}
		}

		result = append(result, data[i])
	}

	return result
}

func validateConfig(cfg *Config) error {
	if cfg.Username != "" && strings.ContainsAny(cfg.Username, "\r\n\t") {
		return fmt.Errorf("username contains invalid characters")
	}

	if cfg.Proxy != "" {
		if err := validateProxyURL(cfg.Proxy); err != nil {
			return fmt.Errorf("proxy: %w", err)
		}
	}

	if cfg.Download != nil {
		if cfg.Download.Parallel > 100 {
			return fmt.Errorf("download.parallel must be between 1 and 100")
		}
		if cfg.Download.OutDir != "" && filepath.IsAbs(cfg.Download.OutDir) {
			if !strings.HasPrefix(cfg.Download.OutDir, "/") && !strings.Contains(cfg.Download.OutDir, ":/") {
				return fmt.Errorf("download.out must be a relative path")
			}
		}
	}

	return nil
}

func validateProxyURL(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	proxyURL = strings.TrimSpace(proxyURL)
	if len(proxyURL) < 8 {
		return fmt.Errorf("proxy URL too short")
	}

	scheme := strings.SplitN(proxyURL, "://", 2)
	if len(scheme) != 2 {
		return fmt.Errorf("proxy URL must have scheme (e.g., http://)")
	}

	validSchemes := map[string]bool{
		"http":  true,
		"https": true,
		"socks5": true,
		"socks5h": true,
	}

	if !validSchemes[strings.ToLower(scheme[0])] {
		return fmt.Errorf("unsupported proxy scheme: %s", scheme[0])
	}

	return nil
}

func mergeConfigWithFlags(cfg *Config, flags *GlobalFlags) {
	if cfg == nil {
		return
	}

	if flags.Username == "" && cfg.Username != "" {
		flags.Username = cfg.Username
	}

	if flags.Password == "" && cfg.Password != "" {
		flags.Password = cfg.Password
	}

	if flags.Proxy == "" && cfg.Proxy != "" {
		flags.Proxy = cfg.Proxy
	}

	if !flags.Verbose && cfg.Verbose {
		flags.Verbose = cfg.Verbose
	}
}
