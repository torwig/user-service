package config

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/torwig/user-service/adapters/user"
	"github.com/torwig/user-service/log"
	"github.com/torwig/user-service/ports/http"
	"github.com/torwig/user-service/ports/http/jwt"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Log        log.Config  `yaml:"log"`
	Repository user.Config `yaml:"repository"`
	JWT        jwt.Config  `yaml:"jwt"`
	HTTP       http.Config `yaml:"http"`
}

func ParseFromFile(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	cfg, err := parse(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configuration")
	}

	return cfg, nil
}

func parse(source io.Reader) (*Config, error) {
	cfg := Config{}

	err := yaml.NewDecoder(source).Decode(&cfg)

	return &cfg, err
}
