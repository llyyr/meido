package utilitymod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

func (m *UtilityMod) StatusLoop() interface{} {

	statusTimer := time.NewTicker(time.Second * 15)

	return func(s *discordgo.Session, r *discordgo.Ready) {
		display := true
		go func() {
			for range statusTimer.C {
				if display {
					memCount := 0
					srvCount := 0
					for _, g := range m.bot.Discord.Guilds() {
						srvCount++
						memCount += g.MemberCount
					}
					s.UpdateStatusComplex(discordgo.UpdateStatusData{
						Activities: []*discordgo.Activity{
							{
								Name: fmt.Sprintf("over %v servers and %v members", srvCount, memCount),
								Type: 3,
							},
						},
					})
				} else {
					s.UpdateStatusComplex(discordgo.UpdateStatusData{
						Activities: []*discordgo.Activity{
							{
								Name: fmt.Sprintf("m?help"),
								Type: discordgo.ActivityTypeGame,
							},
						},
					})
				}
				display = !display
			}
		}()
	}

}
