package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/intrntsrfr/meido/internal/module/search/service"
	iutils "github.com/intrntsrfr/meido/internal/utils"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"go.uber.org/zap"
)

type Module struct {
	*bot.ModuleBase
	search     *service.Service
	imageCache *service.ImageSearchCache
}

func New(b *bot.Bot, logger *zap.Logger) bot.Module {
	logger = logger.Named("Search")
	return &Module{
		ModuleBase: bot.NewModule(b, "Search", logger),
		search:     service.NewService(b.Config.GetString("youtube_token"), b.Config.GetString("open_weather_key")),
		imageCache: service.NewImageSearchCache(),
	}
}

func (m *Module) Hook() error {
	m.Bot.Discord.AddEventHandler(m.imageInteractionHandler)
	return m.RegisterCommands(
		newWeatherCommand(m),
		newYouTubeCommand(m),
		newImageCommand(m),
	)
}

func newWeatherCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "weather",
		Description:      "Finds the weather at a provided location",
		Triggers:         []string{"m?weather"},
		Usage:            "m?weather [city]",
		Cooldown:         0,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}
			query := strings.TrimSpace(strings.Join(msg.RawArgs()[1:], " "))
			if query == "" {
				return
			}
			d, err := m.search.GetWeatherData(query)
			if err != nil {
				_, _ = msg.Reply("I could not find that city :(")
				return
			}
			f := iutils.CelsiusToFahrenheit
			embed := iutils.NewEmbed().
				WithDescription(fmt.Sprintf("Weather in [%v, %v](https://openweathermap.org/city/%v)", d.Name, d.Sys.Country, d.ID)).
				WithOkColor()

			if len(d.Weather) > 0 {
				embed.AddField("⛅ Weather", d.Weather[0].Main, true)
			}
			embed.AddField("🌡️ Temperature", fmt.Sprintf("%.1f°C / %.1f°F", d.Main.Temp, f(d.Main.Temp)), true).
				AddField("Min | Max", fmt.Sprintf("%.1f°C | %.1f°C\n%.1f°F | %.1f°F",
					d.Main.TempMin, d.Main.TempMax, f(d.Main.TempMin), f(d.Main.TempMax)), true).
				AddField("💨 Wind", fmt.Sprintf("%.1f m/s", d.Wind.Speed), true).
				AddField("🌅 Sunrise", fmt.Sprintf("<t:%v:R>", d.Sys.Sunrise), true).
				AddField("🌇 Sunset", fmt.Sprintf("<t:%v:R>", d.Sys.Sunset), true).
				WithFooter("Powered by openweathermap.org", "")
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newYouTubeCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "youtube",
		Description:      "Search for a YouTube video",
		Triggers:         []string{"m?youtube", "m?yt"},
		Usage:            "m?yt [query]",
		Cooldown:         2,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}

			query := strings.TrimSpace(strings.Join(msg.RawArgs()[1:], " "))
			if query == "" {
				return
			}

			ids, err := m.search.SearchYoutube(query)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			if len(ids) < 1 {
				_, _ = msg.Reply("I got no results for that :(")
				return
			}

			_, _ = msg.Reply("https://youtube.com/watch?v=" + ids[0])
		},
	}
}

func newImageCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "image",
		Description:      "Search for an image",
		Triggers:         []string{"m?image", "m?img", "m?im"},
		Usage:            "m?img [query]",
		Cooldown:         2,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}
			query := strings.Join(msg.Args()[1:], " ")
			links, err := m.search.SearchGoogleImages(query)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}
			if len(links) < 1 {
				_, _ = msg.Reply("I found 0 results for that :(")
				return
			}

			embed := iutils.NewEmbed().
				WithTitle("Google Images search results").
				WithOkColor().
				WithImageUrl(links[0]).
				WithFooter(fmt.Sprintf("Image [ %v / %v ]", 1, len(links)), "")

			nextID := uuid.New().String()
			prevID := uuid.New().String()
			stopID := uuid.New().String()

			replyData := &discordgo.MessageSend{
				Components: []discordgo.MessageComponent{
					&discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							&discordgo.Button{
								Label:    "⬅️",
								Style:    discordgo.PrimaryButton,
								CustomID: prevID,
							},
							&discordgo.Button{
								Label:    "➡️",
								Style:    discordgo.PrimaryButton,
								CustomID: nextID,
							},
							&discordgo.Button{
								Label:    "⏹️",
								Style:    discordgo.PrimaryButton,
								CustomID: stopID,
							},
						},
					},
				},
				Reference:       &discordgo.MessageReference{MessageID: msg.ID(), ChannelID: msg.ChannelID(), GuildID: msg.GuildID()},
				AllowedMentions: &discordgo.MessageAllowedMentions{},
				Embed:           embed.Build(),
			}

			reply, err := msg.ReplyComplex(replyData)
			if err != nil {
				return
			}
			searchData := service.NewImageSearch(msg.Message, reply, links, nextID, prevID, stopID)
			m.imageCache.Set(searchData)
			defer func() {
				m.imageCache.Delete(reply.ID)
				reply.Components = nil
				if len(reply.Embeds) > 0 {
					_, _ = msg.Sess.ChannelMessageEditComplex(&discordgo.MessageEdit{
						Components: []discordgo.MessageComponent{},
						ID:         reply.ID,
						Channel:    reply.ChannelID,
						Embed:      reply.Embeds[0],
					})
				}
			}()
			for {
				select {
				case id := <-searchData.UpdateCh:
					switch id {
					case nextID:
						emb := searchData.UpdateEmbed(1)
						_, _ = msg.Sess.ChannelMessageEditEmbed(reply.ChannelID, reply.ID, emb)
					case prevID:
						emb := searchData.UpdateEmbed(-1)
						_, _ = msg.Sess.ChannelMessageEditEmbed(reply.ChannelID, reply.ID, emb)
					case stopID:
						_ = msg.Sess.ChannelMessageDelete(reply.ChannelID, reply.ID)
						_ = msg.Sess.ChannelMessageDelete(msg.ChannelID(), msg.Message.ID)
						m.imageCache.Delete(reply.ID)
						return
					}
				case <-time.After(time.Second * 15):
					return
				}
			}
		},
	}
}

func (m *Module) imageInteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Message == nil || i.Data == nil || i.Data.Type() != discordgo.InteractionMessageComponent {
		return
	}
	msg, ok := m.imageCache.Get(i.Message.ID)
	if !ok {
		return
	}
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: nil,
	})
	if i.GuildID != "" && i.Member.User.ID != msg.AuthorID() {
		return
	}
	if i.GuildID == "" && i.User.ID != msg.AuthorID() {
		return
	}
	msg.UpdateCh <- i.Interaction.MessageComponentData().CustomID
}
