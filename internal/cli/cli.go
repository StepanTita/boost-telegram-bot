package cli

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/boost-telegram-bot/internal/config"
	"github.com/boost-telegram-bot/internal/services/listener"
)

func Run(args []string) bool {
	cfg := config.New(os.Getenv("CONFIG"))
	log := cfg.Logging()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.Error("internal panicked: ", rvr)
		}
	}()

	svc := listener.New(cfg)

	app := &cli.App{
		Commands: cli.Commands{
			{
				Name:  "run",
				Usage: "run bot daemon",
				Action: func(c *cli.Context) error {
					if err := svc.Listen(c.Context); err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(args); err != nil {
		log.Fatal(err, ": service initialization failed")
		return false
	}

	return true
}
