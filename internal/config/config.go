package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

const defaultConfigTemplate = `{
  "username": "admin",
  "password": "changeme",
  "port": "8080",
  "libraries": [
    {
      "name": "My Media",
      "path": "C:/Users/you/Videos"
    }
  ]
}
`

type Config struct {
	Username  string          `json:"username"`
	Password  string          `json:"password"`
	Port      string          `json:"port"`
	Libraries []LibraryConfig `json:"libraries"`
}

type LibraryConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if werr := os.WriteFile(path, []byte(defaultConfigTemplate), 0644); werr != nil {
			log.Fatalf("no config file found and could not create one at %s: %v", path, werr)
		}
		log.Printf("no config found — created a template at %s", path)
		log.Printf("edit it with your username/password, port, and media folder(s), then restart")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) ResolveAddr() string {
	if envAddr := os.Getenv("ADDR"); envAddr != "" {
		return envAddr
	}
	port := strings.TrimSpace(c.Port)
	if port == "" {
		return ":8080"
	}
	if strings.Contains(port, ":") {
		return port
	}
	return ":" + port
}
