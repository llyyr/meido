package mediaconvertmod

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
)

type MediaConvertMod struct {
	*mio.ModuleBase
}

func New() mio.Module {
	return &MediaConvertMod{
		ModuleBase: mio.NewModule("Media"),
	}
}

func (m *MediaConvertMod) Hook() error {
	return m.RegisterPassive(NewJpgLargeConvertPassive(m))
}

func NewJpgLargeConvertPassive(m *MediaConvertMod) *mio.ModulePassive {
	return &mio.ModulePassive{
		Mod:          m,
		Name:         "jpglargeconvert",
		Description:  "Automatically converts jpglarge files to jpg",
		AllowedTypes: mio.MessageTypeCreate,
		Enabled:      true,
		Run:          m.jpglargeconvertPassive,
	}
}

func (m *MediaConvertMod) jpglargeconvertPassive(msg *mio.DiscordMessage) {
	if len(msg.Message.Attachments) < 1 {
		return
	}

	var files []*discordgo.File

	for _, att := range msg.Message.Attachments {
		func() {
			if filepath.Ext(att.URL) != ".jpglarge" {
				return
			}

			res, err := http.Get(att.URL)
			if err != nil {
				return
			}

			if res.StatusCode != 200 {
				return
			}

			defer res.Body.Close()

			files = append(files, &discordgo.File{
				Name:   "converted.jpg",
				Reader: res.Body,
			})
		}()
	}

	if len(files) < 1 {
		return
	}

	msg.Sess.ChannelMessageSendComplex(msg.Message.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("%v, I converted that to JPG for you", msg.Author().Mention()),
		Files:   files,
	})
}

func NewMediaConvertCommand(m *MediaConvertMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "mediaconvert",
		Description:   "Converts some media files from one format to another",
		Triggers:      []string{"m?mediaconvert"},
		Usage:         "m?mediaconvert [url] [target format]",
		Cooldown:      30,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.mediaconvertCommand,
	}
}

func (m *MediaConvertMod) mediaconvertCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 3 {
		return
	}
	// m?mediaconvert link format

	format := msg.Args()[2]
	if format != "mp4" {
		msg.Reply("invalid format")
		return
	}

	src, err := http.Get(msg.RawArgs()[1])
	if err != nil {
		fmt.Println(err)
		msg.Reply("invalid link")
		return
	}
	defer src.Body.Close()

	fmt.Println(src.StatusCode)
	if src.StatusCode != 200 {
		msg.Reply("invalid link")
		return
	}

	p, _ := io.ReadAll(src.Body)
	if http.DetectContentType(p) != "video/webm" {
		msg.Reply("link is not webm")
		return
	}

	if len(p) > 1024*(1024*6) {
		msg.Reply("file too large")
		return
	}

	inp := &bytes.Buffer{}
	inp.Write(p)

	msg.Reply("this might take a while")

	cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "mp4", "-movflags", "frag_keyframe", "pipe:1")
	//cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "mp4", "pipe:1")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	func() {
		defer stdin.Close()
		io.Copy(stdin, inp)
	}()

	var out bytes.Buffer
	io.Copy(&out, stdout)

	cmd.Wait()

	fmt.Println(out.Len())

	if out.Len() > 1024*(1024*8) {
		msg.Reply("file too large")
	}

	msg.Discord.Sess.ChannelFileSend(msg.Message.ChannelID, "video.mp4", &out)
}
