package store

import (
	"errors"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

type Service string

const (
	APIServer    = Service("api")
	SearchServer = Service("search")
)

type Config struct {
	Domain  string
	Flags   []Flag `toml:"Features"`
	Servers []Service
}

func LoadConfig() (Config, error) {
	markStoreLocation, err := GetStoragePath()
	if err != nil {
		return Config{}, errors.Join(errors.New("unable to get mark store location"), err)
	}
	config := &Config{}
	_, err = toml.DecodeFile(path.Join(markStoreLocation, "config.toml"), config)
	if err != nil {
		return Config{}, err
	}
	return *config, err
}

var defaultConfig = Config{
	Domain:  "lukaswerner.com",
	Flags:   []Flag{Embedding},
	Servers: []Service{APIServer},
}

func SaveDefaultConfig() (Config, error) {
	markStoreLocation, err := GetStoragePath()
	if err != nil {
		return defaultConfig, errors.Join(errors.New("unable to get mark store location"), err)
	}
	config := defaultConfig

	f, err := os.Create(path.Join(markStoreLocation, "config.toml"))
	if err != nil {
		return defaultConfig, err
	}
	defer f.Close()

	return defaultConfig, toml.NewEncoder(f).Encode(config)
}
