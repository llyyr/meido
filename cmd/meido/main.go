package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/database"
	"github.com/intrntsrfr/meido/internal/mods/aimod"
	"github.com/intrntsrfr/meido/internal/mods/loggermod"
	"github.com/intrntsrfr/meido/internal/mods/mediaconvertmod"
	"github.com/intrntsrfr/meido/internal/mods/moderationmod"
	"github.com/intrntsrfr/meido/internal/mods/searchmod"
	"github.com/intrntsrfr/meido/internal/mods/testmod"
	"github.com/intrntsrfr/meido/internal/mods/userrolemod"
	"github.com/intrntsrfr/meido/internal/mods/utilitymod"
	"github.com/intrntsrfr/meido/internal/services"
	"github.com/intrntsrfr/owo"
	"github.com/jmoiron/sqlx"
	gogpt "github.com/sashabaranov/go-gpt3"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func main() {

	logger, _ := zap.NewProduction()

	f, err := os.Create("./error_log.dat")
	if err != nil {
		panic("cannot create error log file")
	}
	defer f.Close()
	log.SetOutput(f)

	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("config file not found")
	}
	var config *base.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic("mangled config file, fix it")
	}

	psql, err := sqlx.Connect("postgres", config.ConnectionString)
	if err != nil {
		panic(err)
	}

	db := database.New(psql)
	owoClient := owo.NewClient(config.OwoToken)
	searchService := services.NewSearchService(config.YouTubeToken)
	gptClient := gogpt.NewClient(config.OpenAIToken)

	bot := base.NewBot(config, db, logger.Named("meido"))
	err = bot.Open()
	if err != nil {
		panic(err)
	}

	bot.RegisterMod(testmod.New())
	//bot.RegisterMod(fishmod.New())
	bot.RegisterMod(loggermod.New(config.DmLogChannels))
	bot.RegisterMod(utilitymod.New(bot, db))
	bot.RegisterMod(moderationmod.New(bot, db, logger.Named("moderation")))
	bot.RegisterMod(userrolemod.New(bot, db, owoClient, logger.Named("userrole")))
	bot.RegisterMod(searchmod.New(bot, searchService))
	bot.RegisterMod(mediaconvertmod.New())
	bot.RegisterMod(aimod.New(gptClient, config.GPT3Engine))

	err = bot.Run()
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
