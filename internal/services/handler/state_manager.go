package handler

import (
	"time"

	"github.com/boost-telegram-bot/internal/services"
	"github.com/boost-telegram-bot/internal/utils"
)

type stateManager map[int64]*state

func (c stateManager) touch(chatID int64) {
	if v, ok := c[chatID]; ok {
		v.touch()
	} else {
		c[chatID] = &state{
			modelName:    commandToModelName(services.NeeCommand),
			lastAccessed: utils.CurrentTimestamp(),
		}
	}
}

type state struct {
	modelName    string
	lastAccessed time.Time
}

func (s *state) touch() {
	s.lastAccessed = utils.CurrentTimestamp()
}

func (s *state) setModelName(name string) {
	s.modelName = name
}
