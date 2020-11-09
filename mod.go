package meidov2

type Mod interface {
	Save() error
	Load() error
	Commands() map[string]ModCommand
	Hook(*Bot) error
	RegisterCommand(ModCommand)
	Settings(*DiscordMessage)
	Help(*DiscordMessage)
	Message(*DiscordMessage)
}

type ModCommand interface {
	Name() string
	Description() string
	Triggers() []string
	Usage() string
	Cooldown() int
	RequiredPerms() int
	RequiresOwner() bool
	IsEnabled() bool
	Run(*DiscordMessage)
}

/*
type ModCommand struct {
	Name string
	Aliases []string
	Triggers []string
	RequiredPerms int
	OwnerOnly bool
	Enabled bool
	Run func(*Message)
}
*/
