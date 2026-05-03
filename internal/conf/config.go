package conf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"github.com/spf13/viper"
)

type Server struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	StaticDir string `mapstructure:"static_dir"`
}

type Log struct {
	Level string `mapstructure:"level"`
}

type Database struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

type Config struct {
	Server   Server   `mapstructure:"server"`
	Log      Log      `mapstructure:"log"`
	Database Database `mapstructure:"database"`
}

var AppConfig Config

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func Load(path string) error {
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("json")
		viper.AddConfigPath("data")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix(APP_NAME)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults()

	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %s", viper.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Infof("Config file not found, creating default config")
			if err := os.MkdirAll("data", 0755); err != nil {
				log.Errorf("Failed to create data directory: %v", err)
			}
			if err := viper.SafeWriteConfigAs("data/config.json"); err != nil {
				log.Errorf("Failed to create default config: %v", err)
			}
		} else {
			if fallbackPath, fallbackErr := readConfigWithBOMFallback(path); fallbackErr == nil {
				log.Infof("Using config file with BOM-stripped fallback: %s", fallbackPath)
			} else {
				return fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("unable to decode config into struct: %w", err)
	}
	return nil
}

func readConfigWithBOMFallback(path string) (string, error) {
	configPath := strings.TrimSpace(path)
	if configPath == "" {
		configPath = filepath.Join("data", "config.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	if !bytes.HasPrefix(data, utf8BOM) {
		return "", fmt.Errorf("config file has no UTF-8 BOM prefix")
	}

	configType := strings.TrimPrefix(strings.ToLower(filepath.Ext(configPath)), ".")
	if configType == "" {
		configType = "json"
	}
	viper.SetConfigType(configType)
	if err := viper.ReadConfig(bytes.NewReader(bytes.TrimPrefix(data, utf8BOM))); err != nil {
		return "", err
	}

	return configPath, nil
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 1088)
	viper.SetDefault("server.static_dir", "static/out")
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "data/data.db")
	viper.SetDefault("log.level", "info")
}
