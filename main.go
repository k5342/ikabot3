package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"net/http"
	_ "net/http/pprof"
)

var (
	logger        *zap.Logger
	scheduleStore ScheduleStore
)

// https://discord.com/oauth2/authorize?client_id=1018084105587544166&scope=bot&permissions=10737436672
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger, _ = zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
	go func() {
		logger.Sugar().Info(http.ListenAndServe("localhost:6060", nil))
	}()

	scheduleStore = NewScheduleStore()
	scheduleStore.MaybeRefresh()

	bot, err := LaunchDiscordBot(os.Getenv("IKABOT3_TOKEN"), os.Getenv("IKABOT3_ALLOW_MESSAGE_CONTENT_INTENT") == "TRUE")
	if err != nil {
		logger.Sugar().Errorw("bot creation failed", err)
	}

	logger.Sugar().Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	bot.CloseDiscordBot()
}
