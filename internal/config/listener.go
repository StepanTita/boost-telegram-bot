package config

type Listener interface {
	TelegramApiToken() string
	Debug() bool
}

type listener struct {
	telegramApiToken string
	debug            bool
}

func NewListener(tgApiToken string, debug bool) Listener {
	return &listener{
		telegramApiToken: tgApiToken,
		debug:            debug,
	}
}

func (l listener) TelegramApiToken() string {
	return l.telegramApiToken
}

func (l listener) Debug() bool {
	return l.debug
}
