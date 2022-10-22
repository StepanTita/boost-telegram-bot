package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config interface {
	Logger
	Listener
	Generator
	Templator
}

type config struct {
	Logger
	Listener
	Generator
	Templator
}

type yamlConfig struct {
	LogLevel         string `yaml:"log_level"`
	TelegramApiToken string `yaml:"telegram_api_token"`
	Debug            bool   `yaml:"debug"`
	Generator        struct {
		OpenAIApiToken string            `yaml:"openai_api_token"`
		Models         map[string]string `yaml:"models"`
	} `yaml:"generator"`
	TemplatesDir string `yaml:"templates_dir"`
}

func New(path string) Config {
	cfg := yamlConfig{}

	yamlConfig, err := ioutil.ReadFile(path)
	if err != nil {
		panic(errors.Wrapf(err, "failed to read config %s", path))
	}

	err = yaml.Unmarshal(yamlConfig, &cfg)
	if err != nil {
		panic(errors.Wrapf(err, "failed to unmarshal config %s", path))
	}

	return &config{
		Logger:    NewLogger(cfg.LogLevel),
		Listener:  NewListener(cfg.TelegramApiToken, cfg.Debug),
		Generator: NewGenerator(cfg.Generator.OpenAIApiToken, cfg.Generator.Models),
		Templator: NewTemplator(),
	}
}
