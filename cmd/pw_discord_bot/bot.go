package main

// handles the discord bot integration

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"plane.watch/lib/export"
	"syscall"
)

type (
	PwBot struct {
		pwDiscordBot
		pwWsClient
		pwAlertBot

		log zerolog.Logger
	}
)

func NewPlaneWatchBot(token string) (*PwBot, error) {
	b := PwBot{
		pwAlertBot: pwAlertBot{
			locationUpdates:  make(chan *export.PlaneLocation, 100),
			numUpdateWorkers: 10,
			log:              log.With().Str("Service", "Alert Handler").Logger(),
		},
		pwDiscordBot: pwDiscordBot{
			commands: make(map[string]*discordgo.ApplicationCommand),
			log:      log.With().Str("Service", "Discord Bot").Logger(),
		},
		pwWsClient: pwWsClient{
			wsLog: log.With().Str("Service", "WS Handler").Logger(),
		},
		log: log.With().Str("Service", "PW Main Bot").Logger(),
	}

	b.pwWsClient.handleUpdate = func(update *export.PlaneLocation) {
		b.pwAlertBot.locationUpdates <- update
	}
	b.pwAlertBot.sendAlert = func(pa *proximityAlert) {
		b.pwDiscordBot.sendPlaneAlert(pa)
	}

	if err := b.pwDiscordBot.setup(token); nil != err {
		return nil, err
	}

	return &b, nil
}

func (b *PwBot) Run(c *cli.Context) error {
	b.log.Info().Msg("Starting up...")
	// load our existing alert config
	loadLocationsList()

	go b.handleWebsocketClient(c.String("host"), c.Bool("insecure"))
	b.runAlerts()

	b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		b.log.Info().Msg("Bot is up!")
		b.RegisterCommands(c.Bool("nuke-commands"))

		if err := b.session.UpdateListeningStatus("ADSB"); nil != err {
			b.log.Info().Msgf("Unable to update listening to: %s", err)
		}
	})

	err := b.session.Open()
	if nil != err {
		return err
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	b.log.Info().Msg("Closing...")

	errs := make([]error, 0)
	if err = saveLocationsList(); nil != err {
		errs = append(errs, err)
	}
	if err = b.pwWsClient.stop(); nil != err {
		errs = append(errs, err)
	}
	if err = b.pwAlertBot.stop(); nil != err {
		errs = append(errs, err)
	}
	if err = b.pwDiscordBot.stop(); nil != err {
		errs = append(errs, err)
	}

	if 0 != len(errs) {
		errStr := "Errors when closing:\n"
		for _, err := range errs {
			errStr += err.Error() + ".\n"
			b.log.Error().Err(err).Str("State", "Shutting Down").Send()
		}
		return errors.New(errStr)
	}

	return nil
}
