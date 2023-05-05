package main

import (
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/meido"
	"github.com/intrntsrfr/meido/internal/structs"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger, _ := zap.NewProduction()
	logger = logger.Named("main")

	cfg, err := structs.LoadConfig()
	if err != nil {
		panic(err)
	}

	db, err := database.NewPSQLDatabase(cfg.GetString("connection_string"))
	if err != nil {
		panic(err)
	}

	bot := meido.New(cfg, db, logger.Named("meido"))
	err = bot.Run(true)
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
