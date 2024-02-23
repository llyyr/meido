package fun

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/gol"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"go.uber.org/zap"
)

type module struct {
	*bot.ModuleBase
}

func New(b *bot.Bot, logger *zap.Logger) bot.Module {
	logger = logger.Named("Fun")
	return &module{
		ModuleBase: bot.NewModule(b, "Fun", logger),
	}
}

func (m *module) Hook() error {
	return m.RegisterCommands(newLifeCommand(m))
}

func newLifeCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "life",
		Description:      "Shows a gif of Conway's Game of Life. If no seed is provided, it uses your user ID",
		Triggers:         []string{"m?life"},
		Usage:            "m?life | m?life <seed | user>",
		Cooldown:         5,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			_ = msg.Discord.StartTyping(msg.ChannelID())
			seedStr := msg.AuthorID()
			if len(msg.Args()) > 1 {
				seedStr = strings.Join(msg.Args()[1:], " ")
			}

			buf, seed, err := generateGif(seedStr)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			_, _ = msg.ReplyComplex(&discordgo.MessageSend{
				Content: fmt.Sprintf("Here you go! Seed: `%v`", seed),
				File: &discordgo.File{
					Name:   "game.gif",
					Reader: buf,
				},
				Reference: &discordgo.MessageReference{
					MessageID: msg.Message.ID,
					ChannelID: msg.ChannelID(),
					GuildID:   msg.GuildID(),
				},
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			})
		},
	}
}

func generateGif(seedStr string) (*bytes.Buffer, int64, error) {
	ye := sha1.New()
	_, err := ye.Write([]byte(seedStr))
	if err != nil {
		return nil, 0, err
	}
	seed := int64(binary.BigEndian.Uint64(ye.Sum(nil)[:8]))
	game, err := gol.NewGame(seed, 100, 100, true)
	game.Run(100, 50, false, true, "game.gif", 2)
	buf := bytes.Buffer{}
	_ = game.Export(&buf) // no need to check error, because export will always be populated
	return &buf, seed, nil
}
