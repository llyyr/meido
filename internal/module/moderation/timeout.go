package moderation

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

func newMuteCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "mute",
		Description:      "Mutes a member, making them unable to chat or speak. Duration will be 1 day unless something else is specified.",
		Triggers:         []string{"m?mute"},
		Usage:            "m?mute <user> [duration] | m?mute 163454407999094786 1h30m",
		Cooldown:         1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionModerateMembers,
		CheckBotPerms:    true,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.muteCommand,
	}
}

func (m *module) muteCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}
	duration := time.Hour * 24
	if len(msg.Args()) > 2 {
		pDur, err := time.ParseDuration(msg.Args()[2])
		if err != nil {
			_, _ = msg.Reply("invalid time format - I allow hours and minutes! Example: 1h30m")
			return
		}
		if pDur < time.Minute || pDur > time.Hour*24*28 {
			_, _ = msg.Reply("duration is either too short or too long - Minimum 1 minute, max 28 days")
			return
		}
		duration = pDur
	}
	until := time.Now().Add(duration)

	// get the target member
	targetMember, err := msg.GetMemberAtArg(1)
	if err != nil {
		return
	}

	if msg.AuthorID() == targetMember.User.ID {
		_, _ = msg.Reply("you cannot mute yourself")
		return
	}

	// check if command hierarchy is valid
	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.AuthorID())
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetMember.User.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State().User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		_, _ = msg.Reply("no (you can only mute users who are below you and me in the role hierarchy)")
		return
	}

	// just unmute 4head
	err = msg.Discord.Sess.GuildMemberTimeout(msg.GuildID(), targetMember.User.ID, &until)
	if err != nil {
		_, _ = msg.Reply("I was unable to mute that member")
		return
	}
	_, _ = msg.Reply(fmt.Sprintf("%v has been timed out for %v", targetMember.User, duration))
}

func newUnmuteCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "unmute",
		Description:      "Unmutes a member",
		Triggers:         []string{"m?unmute"},
		Usage:            "m?unmute <user>",
		Cooldown:         1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionModerateMembers,
		CheckBotPerms:    true,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.unmuteCommand,
	}
}

func (m *module) unmuteCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}
	// get the target member
	targetMember, err := msg.GetMemberAtArg(1)
	if err != nil {
		return
	}
	if targetMember.CommunicationDisabledUntil == nil {
		return
	}
	if msg.AuthorID() == targetMember.User.ID {
		return
	}
	if msg.TargetRoleIsLowest(targetMember) {
		_, _ = msg.Reply("no (you can only unmute users who are below you and me in the role hierarchy)")
		return
	}

	// just unmute 4head
	err = msg.Discord.Sess.GuildMemberTimeout(msg.GuildID(), targetMember.User.ID, nil)
	if err != nil {
		_, _ = msg.Reply("I was unable to unmute that member")
		return
	}
	_, _ = msg.Reply(fmt.Sprintf("unmuted %v", targetMember.User))
}
