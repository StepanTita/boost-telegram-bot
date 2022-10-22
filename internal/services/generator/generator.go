package generator

import (
	"context"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/boost-telegram-bot/internal/config"
)

const (
	promptTerminator     = "\\n\\n###\\n\\n"
	completionTerminator = " END"
)

type Generator interface {
	Generate(ctx context.Context, modelName string, keywords []string) (string, error)
}

type generator struct {
	client gpt3.Client

	models map[string]string

	cfg config.Config
	log *logrus.Entry
}

func New(cfg config.Config) Generator {
	client := gpt3.NewClient(cfg.OpenAIApiToken())
	return &generator{
		client: client,
		models: cfg.Models(),

		cfg: cfg,
		log: cfg.Logging().WithField("service", "generator"),
	}
}

func (g generator) Generate(ctx context.Context, modelName string, keywords []string) (string, error) {
	g.log.WithFields(logrus.Fields{
		"model_name": modelName,
		"prompt":     strings.Join(keywords, ",") + promptTerminator,
		"max_tokens": 256,
		"stop":       completionTerminator,
	}).Debug("Starting generate...")
	resp, err := g.client.CompletionWithEngine(ctx, g.models[modelName], gpt3.CompletionRequest{
		Prompt:    []string{strings.Join(keywords, ",") + promptTerminator},
		MaxTokens: gpt3.IntPtr(256),
		Stop:      []string{completionTerminator},
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to make completion request")
	}

	if len(resp.Choices) < 1 {
		return "", errors.New("empty choices unexpected")
	}

	return resp.Choices[0].Text, nil
}
