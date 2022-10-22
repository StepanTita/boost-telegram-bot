package listener

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/boost-telegram-bot/internal/config"
	"github.com/boost-telegram-bot/internal/services"
	"github.com/boost-telegram-bot/internal/services/handler"
	"github.com/boost-telegram-bot/internal/utils"
)

type Listener interface {
	Listen(ctx context.Context) error
}

type listener struct {
	handler handler.Handler

	cfg config.Config
	log *logrus.Entry
}

func New(cfg config.Config) Listener {
	return &listener{
		handler: handler.New(cfg),

		cfg: cfg,
		log: cfg.Logging().WithField("service", "listener"),
	}
}

// Listen TODO: make value receiver when states is eliminated
func (l listener) Listen(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(l.cfg.TelegramApiToken())
	if err != nil {
		return errors.Wrap(err, "failed to initialize bot API")
	}

	l.log.Info("Running refresh state...")
	go l.handler.RefreshState()

	l.log.Infof("Authorized on account %s", bot.Self.UserName)

	buttonsConfig, updates, err := l.configUpdates(bot)
	if err != nil {
		return errors.Wrap(err, "failed to configurate bot")
	}

	for update := range updates {
		log := l.log.WithField("update_id", update.UpdateID)
		log.Debug("reading updates...")
		if update.Message == nil {
			continue
		}

		// refresh the state
		l.handler.Touch(update.Message.Chat.ID)

		log.Debugf("update from chat: %d, with message: %s", update.Message.Chat.ID, update.Message.Text)

		var msg tgbotapi.MessageConfig

		if update.Message.IsCommand() {
			log.Debug("handling command...")
			msg = l.handler.HandleCommand(ctx, update.Message)
			log.Debugf("command handled with output: %s", msg.Text)
		} else {
			log.Debug("handling message...")
			msg, err = l.handler.HandleMessage(ctx, update.Message)
			if err != nil {
				log.WithError(err).Error("failed to handle message")
			}
			log.Debugf("message handled with output: %s", msg.Text)
		}

		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = buttonsConfig

		if _, err := bot.Send(msg); err != nil {
			log.WithError(err).Error("failed to send message to bot API")
		}
	}
	return nil
}

func (l listener) configUpdates(bot *tgbotapi.BotAPI) (tgbotapi.ReplyKeyboardMarkup, tgbotapi.UpdatesChannel, error) {
	if err := tgbotapi.SetLogger(l.log.WithField("[BOT]", bot.Self.UserName)); err != nil {
		return tgbotapi.ReplyKeyboardMarkup{}, nil, errors.Wrap(err, "failed to set bog logger")
	}
	bot.Debug = l.cfg.Debug()

	commandsConfig := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     services.StartCommand,
			Description: "start the interaction",
		}, tgbotapi.BotCommand{
			Command:     services.NeeCommand,
			Description: "set current model as 'Named Entity Extraction'",
		}, tgbotapi.BotCommand{
			Command:     services.PsCommand,
			Description: "set current model as 'Punctuation Stopwords'",
		},
		tgbotapi.BotCommand{
			Command:     services.PslCommand,
			Description: "set current model as 'Punctuation Stopwords Lemmatization'",
		},
	)

	buttonsConfig := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(utils.ToCommand(services.StartCommand)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(utils.ToCommand(services.NeeCommand)),
			tgbotapi.NewKeyboardButton(utils.ToCommand(services.PsCommand)),
			tgbotapi.NewKeyboardButton(utils.ToCommand(services.PslCommand)),
		),
	)

	updateConfig := tgbotapi.NewUpdate(0)

	updateConfig.Timeout = 30

	_, err := bot.Request(commandsConfig)
	if err != nil {
		return tgbotapi.ReplyKeyboardMarkup{}, nil, errors.Wrap(err, "failed to send commands config")
	}

	// Start polling Telegram for updates.
	return buttonsConfig, bot.GetUpdatesChan(updateConfig), err
}
