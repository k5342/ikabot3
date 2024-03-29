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

type ModeInfo struct {
	Mode
	ModeName   string
	Identifier string
	Color      int
}

type Mode interface {
	getModeName() string
	getIdentifier() string
	getColor() int
}

func (mi ModeInfo) getModeName() string {
	return mi.ModeName
}

func (mi ModeInfo) getIdentifier() string {
	return mi.Identifier
}

func (mi ModeInfo) getColor() int {
	return mi.Color
}

var ModeTable map[string]ModeInfo

func init() {
	ModeTable = map[string]ModeInfo{
		"OPEN": {
			ModeName:   "バンカラマッチ（オープン）",
			Identifier: "OPEN",
			Color:      0xf64a10,
		},
		"CHALLENGE": {
			ModeName:   "バンカラマッチ（チャレンジ）",
			Identifier: "CHALLENGE",
			Color:      0xf64a10,
		},
		"X": {
			ModeName:   "Xマッチ",
			Identifier: "X",
			Color:      0x74f1a2,
		},
		"SALMON": {
			ModeName:   "サーモンラン",
			Identifier: "SALMON",
			Color:      0xff501e,
		},
		"BIGRUN": {
			ModeName:   "ビッグラン",
			Identifier: "SALMON",
			Color:      0xfe0de8,
		},
		"REGULAR": {
			ModeName:   "レギュラーマッチ",
			Identifier: "REGULAR",
			Color:      0xd0f623,
		},
	}
}

func getMode(identifier string) Mode {
	mode, found := ModeTable[identifier]
	if found {
		return mode
	} else {
		return ModeInfo{
			ModeName:   "",
			Identifier: identifier,
			Color:      0x0,
		}
	}
}

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
