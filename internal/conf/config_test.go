package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func resetViperForTest() {
	viper.Reset()
	AppConfig = Config{}
}

func TestLoadSupportsUTF8BOMConfig(t *testing.T) {
	resetViperForTest()
	defer resetViperForTest()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	config := []byte("{\n  \"server\": {\n    \"host\": \"127.0.0.1\",\n    \"port\": 18080\n  },\n  \"database\": {\n    \"type\": \"sqlite\",\n    \"path\": \"" + filepath.ToSlash(filepath.Join(tempDir, "octopus.db")) + "\"\n  },\n  \"log\": {\n    \"level\": \"debug\"\n  }\n}\n")
	payload := append(append([]byte{}, utf8BOM...), config...)
	if err := os.WriteFile(configPath, payload, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := Load(configPath); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if AppConfig.Server.Host != "127.0.0.1" {
		t.Fatalf("server.host = %q, want %q", AppConfig.Server.Host, "127.0.0.1")
	}
	if AppConfig.Server.Port != 18080 {
		t.Fatalf("server.port = %d, want 18080", AppConfig.Server.Port)
	}
	if AppConfig.Database.Type != "sqlite" {
		t.Fatalf("database.type = %q, want sqlite", AppConfig.Database.Type)
	}
	if AppConfig.Log.Level != "debug" {
		t.Fatalf("log.level = %q, want debug", AppConfig.Log.Level)
	}
}
