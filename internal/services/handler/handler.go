package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/boost-telegram-bot/internal/config"
	"github.com/boost-telegram-bot/internal/services"
	"github.com/boost-telegram-bot/internal/services/generator"
	"github.com/boost-telegram-bot/internal/utils"
)

type Handler interface {
	RefreshState()
	Touch(chatID int64)

	HandleMessage(ctx context.Context, incomingMsg *tgbotapi.Message) (tgbotapi.MessageConfig, error)
	HandleCommand(ctx context.Context, incomingMsg *tgbotapi.Message) tgbotapi.MessageConfig
}

type handler struct {
	generator generator.Generator
	// TODO: ELIMINATE THIS AS FAST AS POSSIBLE, NEED !!!ONLY!!! FOR TESTING PURPOSES
	states stateManager // chatID:states

	log       *logrus.Entry
	templator config.Templator
}

func New(cfg config.Config) Handler {
	return &handler{
		generator: generator.New(cfg),

		states: make(stateManager),

		log:       cfg.Logging().WithField("service", "handler"),
		templator: cfg,
	}
}

func (h *handler) RefreshState() {
	i := 0
	if err := utils.RunEvery(5*time.Minute, func() error {
		h.log.WithField("iteration", i).Info("Refreshing the state...")
		toRemoveIDs := make([]int64, 0, 10)
		for chatID, st := range h.states {
			if utils.CurrentTimestamp().Sub(st.lastAccessed) >= 15*time.Minute {
				toRemoveIDs = append(toRemoveIDs, chatID)
			}
		}
		for _, removeID := range toRemoveIDs {
			delete(h.states, removeID)
		}
		i++
		return nil
	}); err != nil {
		panic(errors.Wrap(err, "failed to refresh states"))
	}
}

func (h *handler) HandleCommand(_ context.Context, incomingMsg *tgbotapi.Message) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(incomingMsg.Chat.ID, "")

	switch incomingMsg.Command() {
	case services.StartCommand:
		msg.Text = fmt.Sprintf(h.templator.Template(services.StartCommand), incomingMsg.From.UserName)
	case services.NeeCommand:
		h.states[incomingMsg.Chat.ID].setModelName(commandToModelName(services.NeeCommand))
		msg.Text = fmt.Sprintf(h.templator.Template(services.NeeCommand), commandToModelName(services.NeeCommand))
	case services.PslCommand:
		h.states[incomingMsg.Chat.ID].setModelName(commandToModelName(services.PslCommand))
		msg.Text = fmt.Sprintf(h.templator.Template(services.PslCommand), commandToModelName(services.PslCommand))
	case services.PsCommand:
		h.states[incomingMsg.Chat.ID].setModelName(commandToModelName(services.PsCommand))
		msg.Text = fmt.Sprintf(h.templator.Template(services.PsCommand), commandToModelName(services.PsCommand))
	default:
		msg.Text = "I don't know that command"
	}
	return msg
}

func (h *handler) HandleMessage(ctx context.Context, incomingMsg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMsg.Chat.ID, "")

	keywords := utils.Map(strings.Split(incomingMsg.Text, ","), strings.TrimSpace)

	output, err := h.generator.Generate(ctx, h.states[incomingMsg.Chat.ID].modelName, keywords)
	if err != nil {
		return tgbotapi.MessageConfig{}, errors.Wrap(err, "failed to generate prediction")
	}

	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.Text = fmt.Sprintf("%s _%s_", "📰", tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, output))
	return msg, nil
}

func (h *handler) Touch(chatID int64) {
	h.states.touch(chatID)
}

func commandToModelName(command string) string {
	switch command {
	case services.NeeCommand:
		return "named_entity_extraction"
	case services.PsCommand:
		return "punctuation_stopwords"
	case services.PslCommand:
		return "punctuation_stopwords_lemmatization"
	default:
		panic("unknown command to model conversion")
	}
}
