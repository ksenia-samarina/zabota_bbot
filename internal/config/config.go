package config

import (
	"github.com/ksenia-samarina/zabota_bbot/internal/logger"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const configFile = "internal/config/config.yaml"

type Config struct {
	Token string `yaml:"token"` // Токен бота в телеграме.
}

type Service struct {
	config Config
}

func New() (*Service, error) {
	s := &Service{}

	rawYAML, err := os.ReadFile(configFile)
	if err != nil {
		logger.Error("Ошибка reading config file", "err", err)
		return nil, errors.Wrap(err, "reading config file")
	}

	err = yaml.Unmarshal(rawYAML, &s.config)
	if err != nil {
		logger.Error("Ошибка parsing yaml", "err", err)
		return nil, errors.Wrap(err, "parsing yaml")
	}

	return s, nil
}

func (s *Service) Token() string {
	return s.config.Token
}

func (s *Service) GetConfig() Config {
	return s.config
}
