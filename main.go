package main

import (
	"fmt"
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
	defer logger.Sync()
	go func() {
		logger.Sugar().Info(http.ListenAndServe("localhost:6060", nil))
	}()

	scheduleStore = NewScheduleStore()
	scheduleStore.MaybeRefresh()

	cmds := []string{
		"次の次の前の次の次のガチマッチ",
		"ガチマ",
		"次のガチマ",
		"次のオープンマッチ",
		"ガチマアサリ",
		"次のリグマヤグラ",
		"次のナワバリバトル",
		"エリア20",
		"19 時のガチマッチ",
		"ガチマ 20",
		"次のエリア"}
	for _, cmd := range cmds {
		query := Parse(cmd)
		fmt.Printf("%#v\n", query)
		fmt.Printf("%#v\n", scheduleStore.Search(query))
	}

	bot, err := LaunchDiscordBot(os.Getenv("IKABOT3_TOKEN"))
	if err != nil {
		logger.Sugar().Errorw("bot creation failed", err)
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	bot.CloseDiscordBot()
}
