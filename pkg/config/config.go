package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 代表全局配置文件结构，与 config.yaml 对齐。
// 字段保持首字母大写以便 yaml 解码。
type Config struct {
	PluginDir   string `yaml:"plugin_dir"`
	DefaultFlow string `yaml:"default_flow"`
	LogLevel    string `yaml:"log_level"`
	LogDev      bool   `yaml:"log_dev"`

	DB struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`

	Queue struct {
		Impl   string `yaml:"impl"`
		Addr   string `yaml:"addr"`
		Stream string `yaml:"stream"`
	} `yaml:"queue"`
}

// Load 从 path 读取 yaml，如 path 为空则默认 ./config.yaml。
func Load(path string) (*Config, error) {
	if path == "" {
		path = "./config.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// 填补默认值
	if cfg.PluginDir == "" {
		cfg.PluginDir = "./plugins"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	// 默认非开发模式
	
	return &cfg, nil
}
