package app

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server   *ServerConfig
	MongoUrl string
}

type ServerConfig struct {
	Address string
	Port    int
}

type RedisConfig struct {
	Addr string
	Pwd  string
}

func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
