package discord

import (
	"io"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordInteraction struct {
	Sess         DiscordSession `json:"-"`
	Discord      *Discord       `json:"-"`
	Interaction  *discordgo.Interaction
	TimeReceived time.Time
	Shard        int
}

func (it *DiscordInteraction) ID() string {
	return it.Interaction.ID
}

func (it *DiscordInteraction) ChannelID() string {
	return it.Interaction.ChannelID
}

func (it *DiscordInteraction) AuthorID() string {
	if it.Interaction.GuildID == "" {
		return it.Interaction.User.ID
	}
	return it.Interaction.Member.User.ID
}

func (it *DiscordInteraction) IsDM() bool {
	return it.Interaction.GuildID == ""
}

func (it *DiscordInteraction) Respond(text string) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) RespondEmbed(embed *discordgo.MessageEmbed) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) RespondEphemeral(text string) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) RespondFile(text, name string, reader io.Reader) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
			Files: []*discordgo.File{{
				Name:   name,
				Reader: reader,
			}},
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

type DiscordApplicationCommand struct {
	*DiscordInteraction
	Data    discordgo.ApplicationCommandInteractionData
	options map[string]*discordgo.ApplicationCommandInteractionDataOption
}

func (d *DiscordApplicationCommand) Name() string {
	return d.Data.Name
}

// Options returns a *discordgo.ApplicationCommandInteractionDataOption given
// by key. 
func (d *DiscordApplicationCommand) Options(key string) (*discordgo.ApplicationCommandInteractionDataOption, bool) {
	if d.options == nil {
		d.options = flattenOptions(d.Data.Options)
	}
	val, ok := d.options[key]
	return val, ok
}

func flattenOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	result := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	flattenOptionsImpl(options, result)
	return result
}

func flattenOptionsImpl(options []*discordgo.ApplicationCommandInteractionDataOption, result map[string]*discordgo.ApplicationCommandInteractionDataOption) {
	for _, option := range options {
		if option.Options != nil {
			flattenOptionsImpl(option.Options, result)
		} else {
			result[option.Name] = option
		}
	}
}

type DiscordMessageComponent struct {
	*DiscordInteraction
	Data discordgo.MessageComponentInteractionData
}

type DiscordModalSubmit struct {
	*DiscordInteraction
	Data discordgo.ModalSubmitInteractionData
}
