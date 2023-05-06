package meido

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/module/utility"
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Meido struct {
	Bot    *mio.Bot
	logger *zap.Logger
}

func New(config mio.Configurable, db database.DB, log *zap.Logger) *Meido {
	bot := mio.NewBot(config, db, log.Named("mio"))

	//bot.RegisterModule(administration.New(bot, logger))
	//bot.RegisterModule(testing.New(bot, logger))
	//bot.RegisterModule(fun.New(bot, logger))
	//bot.RegisterModule(fishmod.New())
	bot.RegisterModule(utility.New(bot, db, log))
	//bot.RegisterModule(moderation.New(bot, db, logger))
	//bot.RegisterModule(customrole.New(bot, db, logger))
	//bot.RegisterModule(search.New(bot, logger))
	//bot.RegisterModule(mediaconvertmod.New())
	//bot.RegisterModule(aimod.New(gptClient, config.GPT3Engine))

	return &Meido{
		Bot:    bot,
		logger: log,
	}
}

func (m *Meido) Run(useDefHandlers bool) error {
	if err := m.Bot.Open(useDefHandlers); err != nil {
		return err
	}

	// register modules here
	// register mio event handlers here
	m.registerMioHandlers()
	// register discord event handlers here
	m.registerDiscordHandlers()

	if err := m.Bot.Run(); err != nil {
		return err
	}
	return nil
}

func (m *Meido) Close() {
	m.Bot.Close()
}

func (m *Meido) registerMioHandlers() {
	m.Bot.AddEventHandler("command_ran", logCommand(m))
	m.Bot.AddEventHandler("command_panicked", logCommandPanicked(m))
}

func logCommand(m *Meido) func(i interface{}) {
	return func(i interface{}) {
		cmd, _ := i.(*mio.CommandRan)
		err := m.Bot.DB.CreateCommandLogEntry(&structs.CommandLogEntry{
			Command:   cmd.Command.Name,
			Args:      strings.Join(cmd.Message.Args(), " "),
			UserID:    cmd.Message.AuthorID(),
			GuildID:   cmd.Message.GuildID(),
			ChannelID: cmd.Message.ChannelID(),
			MessageID: cmd.Message.Message.ID,
			SentAt:    time.Now(),
		})
		if err != nil {
			m.logger.Error("error logging command", zap.Error(err))
		}
	}
}

func logCommandPanicked(m *Meido) func(i interface{}) {
	return func(i interface{}) {
		cmd, _ := i.(*mio.CommandPanicked)
		m.logger.Error("command panicked",
			zap.Any("command", cmd.Command),
			zap.Any("message", cmd.Message),
			zap.String("stack trace", cmd.StackTrace),
		)
	}
}

func (m *Meido) registerDiscordHandlers() {
	m.Bot.Discord.AddEventHandler(insertGuild(m))
}

func insertGuild(m *Meido) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if _, err := m.Bot.DB.GetGuild(g.Guild.ID); err != nil && err == sql.ErrNoRows {
			if err = m.Bot.DB.CreateGuild(g.Guild.ID); err != nil {
				m.logger.Error("could not create new guild", zap.Error(err), zap.String("guild ID", g.ID))
			}
		}
	}
}
