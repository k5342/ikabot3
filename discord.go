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

func createSingleStageInfoEmbed(sr SearchResult) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: sr.Slot1.Rule.Name,
		Author: &discordgo.MessageEmbedAuthor{
			Name: printAsReadableName(sr.Query.Mode),
		},
		Description: fmt.Sprintf("%d時～%d時\n\n%s\n%s",
			sr.Slot1.StartTime.Hour(), sr.Slot1.EndTime.Hour(),
			sr.Slot1.Stages[0].Name, sr.Slot1.Stages[1].Name),
	}
	return embed
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

func createTwoStageInfoEmbeds(sr SearchResult) []*discordgo.MessageEmbed {
	embed1 := &discordgo.MessageEmbed{
		Title: sr.Slot1.Rule.Name,
		Author: &discordgo.MessageEmbedAuthor{
			Name: printAsReadableName("CHALLENGE"),
		},
		Description: fmt.Sprintf("%d時～%d時\n\n%s\n%s",
			sr.Slot1.StartTime.Hour(), sr.Slot1.EndTime.Hour(),
			sr.Slot1.Stages[0].Name, sr.Slot1.Stages[1].Name),
	}
	embed2 := &discordgo.MessageEmbed{
		Title: sr.Slot2.Rule.Name,
		Author: &discordgo.MessageEmbedAuthor{
			Name: printAsReadableName("OPEN"),
		},
		Description: fmt.Sprintf("%d時～%d時\n\n%s\n%s",
			sr.Slot2.StartTime.Hour(), sr.Slot2.EndTime.Hour(),
			sr.Slot2.Stages[0].Name, sr.Slot2.Stages[1].Name),
	}
	return []*discordgo.MessageEmbed{embed1, embed2}
}

func isMentioned(user *discordgo.User, mentions []*discordgo.User) bool {
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
	regex := regexp.MustCompile(` *<@\d+?> *`)
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
		if isMentioned(s.State.User, m.Mentions) {
			_, err = s.ChannelMessageSendReply(m.ChannelID, "Not Found!", m.Reference())
		}
	}
	fmt.Println(err)
}
