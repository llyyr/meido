package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/meido"
	"github.com/intrntsrfr/meido/internal/structs"
)

func main() {
	cfg, err := structs.LoadConfig()
	if err != nil {
		panic(err)
	}

	db, err := database.NewPSQLDatabase(cfg.GetString("connection_string"))
	if err != nil {
		panic(err)
	}

	bot := meido.New(cfg, db)
	err = bot.Run(context.Background(), true)
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
