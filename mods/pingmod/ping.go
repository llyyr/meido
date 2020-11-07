package pingmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"time"
)

type PingMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
}

func New() meidov2.Mod {
	return &PingMod{}
}

func (m *PingMod) Save() error {
	return nil
}

func (m *PingMod) Load() error {
	return nil
}

func (m *PingMod) Settings(msg *meidov2.DiscordMessage) {

}
func (m *PingMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *PingMod) Commands() []meidov2.ModCommand {
	return nil
}

func (m *PingMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println(len(r.Guilds))
		fmt.Println(r.User.String())
	})

	m.commands = append(m.commands, m.PingCommand)
	//m.commands = append(m.commands, m.check)

	return nil
}

func (m *PingMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}

func (m *PingMod) PingCommand(msg *meidov2.DiscordMessage) {
	if msg.Message.Content != "m?ping" {
		return
	}

	m.cl <- msg

	startTime := time.Now()

	first, err := msg.Reply("Ping")
	if err != nil {
		return
	}

	now := time.Now()
	discordLatency := now.Sub(startTime)
	botLatency := now.Sub(msg.TimeReceived)

	msg.Sess.ChannelMessageEdit(msg.Message.ChannelID, first.ID,
		fmt.Sprintf("Pong!\nDiscord delay: %s\nBot delay: %s", discordLatency, botLatency))
}
