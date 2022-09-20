package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	Session                   *discordgo.Session
	AllowMessageContentIntent bool
}

func LaunchDiscordBot(botToken string, allowMessageContentIntent bool) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}
	dg.AddHandler(messageCreate)
	dg.Identify.Intents |= discordgo.IntentsGuildMessages
	if allowMessageContentIntent {
		dg.Identify.Intents |= discordgo.IntentMessageContent
	}
	err = dg.Open()
	if err != nil {
		return nil, err
	}

	return &DiscordBot{
		Session:                   dg,
		AllowMessageContentIntent: allowMessageContentIntent,
	}, nil
}

func (bot *DiscordBot) CloseDiscordBot() {
	bot.Session.Close()
}

func printAsReadableName(mode string) string {
	if mode == "REGULAR" {
		return "レギュラーマッチ"
	}
	if mode == "CHALLENGE" {
		return "バンカラマッチ（チャレンジ）"
	}
	if mode == "OPEN" {
		return "バンカラマッチ（オープン）"
	}
	return ""
}

func createMessageEmbedFromTimeSlotInfo(tsi *TimeSlotInfo, modeLabel string) *discordgo.MessageEmbed {
	if tsi == nil {
		return &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name: printAsReadableName(modeLabel),
			},
			Description: "Not Found!",
		}
	} else {
		return &discordgo.MessageEmbed{
			Title: tsi.Rule.Name,
			Author: &discordgo.MessageEmbedAuthor{
				Name: printAsReadableName(modeLabel),
			},
			Description: fmt.Sprintf("%d/%d %d時～%d/%d %d時\n\n%s\n%s",
				tsi.StartTime.Month(), tsi.StartTime.Day(), tsi.StartTime.Hour(),
				tsi.EndTime.Month(), tsi.EndTime.Day(), tsi.EndTime.Hour(),
				tsi.Stages[0].Name, tsi.Stages[1].Name),
		}
	}
}

func createSingleStageInfoEmbed(sr SearchResult) *discordgo.MessageEmbed {
	return createMessageEmbedFromTimeSlotInfo(sr.Slot1, sr.Query.Mode)
}

func createTwoStageInfoEmbeds(sr SearchResult) []*discordgo.MessageEmbed {
	embed1 := createMessageEmbedFromTimeSlotInfo(sr.Slot1, "CHALLENGE")
	embed2 := createMessageEmbedFromTimeSlotInfo(sr.Slot2, "OPEN")
	return []*discordgo.MessageEmbed{embed1, embed2}
}

func isMentioned(user *discordgo.User, mentions []*discordgo.User, messageContent string) bool {
	for _, mention := range mentions {
		if mention.ID == user.ID {
			return true
		}
	}

	return false
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	input := m.Message.Content
	if input == "" {
		return
	}

	// remove mention syntax
	regex := regexp.MustCompile(` *<@&?\d+?> *`)
	input = regex.ReplaceAllString(input, "")
	// remove spaces
	input = strings.ReplaceAll(input, " ", "")

	// parse
	query := Parse(input)

	// ignore when no match
	if query.OriginalText == "" {
		return
	}

	// query
	scheduleStore.MaybeRefresh()
	sr := scheduleStore.Search(query)

	// reply
	var err error
	if sr.Found {
		if sr.IsTwoSlots {
			_, err = s.ChannelMessageSendEmbedsReply(m.ChannelID, createTwoStageInfoEmbeds(sr), m.Reference())
		} else {
			_, err = s.ChannelMessageSendEmbedReply(m.ChannelID, createSingleStageInfoEmbed(sr), m.Reference())
		}
	} else {
		if isMentioned(s.State.User, m.Mentions, input) {
			_, err = s.ChannelMessageSendReply(m.ChannelID, "Not Found!", m.Reference())
		}
	}
	if err != nil {
		logger.Sugar().Error(err)
	}
}
