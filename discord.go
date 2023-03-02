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
	registeredCommands        []*discordgo.ApplicationCommand
}

func LaunchDiscordBot(botToken string, allowMessageContentIntent bool) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}
	dg.AddHandler(messageCreate)
	dg.AddHandler(interactionCreate)
	dg.Identify.Intents |= discordgo.IntentsGuildMessages
	if allowMessageContentIntent {
		dg.Identify.Intents |= discordgo.IntentMessageContent
	}
	err = dg.Open()
	if err != nil {
		return nil, err
	}

	bot := DiscordBot{
		Session:                   dg,
		AllowMessageContentIntent: allowMessageContentIntent,
	}
	bot.setupSlashCommands()

	return &bot, nil
}

func (bot *DiscordBot) CloseDiscordBot() {
	for _, val := range bot.registeredCommands {
		err := bot.Session.ApplicationCommandDelete(bot.Session.State.User.ID, "", val.ID)
		if err == nil {
			logger.Sugar().Infof("Deleted a command '%#v'", val.Name)
		} else {
			logger.Sugar().Errorf("Cannot delete command '%v': %v", val.Name, err)
		}
	}
	bot.Session.Close()
}

func (bot *DiscordBot) setupSlashCommands() {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "regular",
			Description: "Return a schedule for regular match",
		},
		{
			Name:        "bankara",
			Description: "Return a schedule for both Open and Challenge match",
		},
		{
			Name:        "open",
			Description: "Return a schedule for Open match",
		},
		{
			Name:        "challenge",
			Description: "Return a schedule for Challenge match",
		},
		{
			Name:        "salmon",
			Description: "Return a schedule for Salmon Run",
		},
		{
			Name:        "x",
			Description: "Return a schedule for X Match",
		},
		{
			Name:        "rule",
			Description: "Search both schedules from Open and Challenge match by rule name",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "rule",
					Description: "a rule name to search",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name: "turf-war",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Japanese: "ナワバリバトル",
							},
							Value: "TURF_WAR",
						},
						{
							Name: "area",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Japanese: "ガチエリア",
							},
							Value: "AREA",
						},
						{
							Name: "rainmarker",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Japanese: "ガチホコバトル",
							},
							Value: "GOAL",
						},
						{
							Name: "tower-control",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Japanese: "ガチヤグラ",
							},
							Value: "LOFT",
						},
						{
							Name: "clam-blitz",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Japanese: "ガチアサリ",
							},
							Value: "CLAM",
						},
					},
				},
			},
		},
	}
	bot.registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for idx, val := range commands {
		registered, err := bot.Session.ApplicationCommandCreate(bot.Session.State.User.ID, "", val)
		if err == nil {
			logger.Sugar().Infof("Created a command '%#v'", val.Name)
		} else {
			logger.Sugar().Errorf("Cannot create command '%#v': %#v", val.Name, err)
		}
		bot.registeredCommands[idx] = registered
	}
}

func printWeaponsList(weapons []WeaponInfo) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s", weapons[0].Name, weapons[1].Name, weapons[2].Name, weapons[3].Name)
}

func createMessageEmbedFromTimeSlotInfo(tsi *TimeSlotInfo, mode Mode) *discordgo.MessageEmbed {
	if tsi == nil {
		return &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name: mode.getModeName(),
			},
			Description: "Not Found!",
		}
	}
	if mode.getIdentifier() == "SALMON" {
		return &discordgo.MessageEmbed{
			Title: tsi.Stage.Name,
			Author: &discordgo.MessageEmbedAuthor{
				Name: mode.getModeName(),
			},
			Description: fmt.Sprintf("%d/%d %d時～%d/%d %d時\n\n%s",
				tsi.StartTime.Month(), tsi.StartTime.Day(), tsi.StartTime.Hour(),
				tsi.EndTime.Month(), tsi.EndTime.Day(), tsi.EndTime.Hour(),
				printWeaponsList(tsi.Weapons)),
			Color: mode.getColor(),
		}
	} else {
		return &discordgo.MessageEmbed{
			Title: tsi.Rule.Name,
			Author: &discordgo.MessageEmbedAuthor{
				Name: mode.getModeName(),
			},
			Description: fmt.Sprintf("%d/%d %d時～%d/%d %d時\n\n%s\n%s",
				tsi.StartTime.Month(), tsi.StartTime.Day(), tsi.StartTime.Hour(),
				tsi.EndTime.Month(), tsi.EndTime.Day(), tsi.EndTime.Hour(),
				tsi.Stages[0].Name, tsi.Stages[1].Name),
			Color: mode.getColor(),
		}
	}
}

func createSingleStageInfoEmbed(sr SearchResult) *discordgo.MessageEmbed {
	return createMessageEmbedFromTimeSlotInfo(sr.Slots[0], sr.Query.Mode)
}

func createTwoStageInfoEmbeds(sr SearchResult) []*discordgo.MessageEmbed {
	embed1 := createMessageEmbedFromTimeSlotInfo(sr.Slots[0], getMode("CHALLENGE"))
	embed2 := createMessageEmbedFromTimeSlotInfo(sr.Slots[1], getMode("OPEN"))
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
		if len(sr.Slots) > 1 {
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

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName2mode := map[string]string{
		"regular":   "REGULAR",
		"bankara":   "BANKARA",
		"open":      "OPEN",
		"challenge": "CHALLENGE",
		"salmon":    "SALMON",
		"x":         "X",
	}
	commandName := i.ApplicationCommandData().Name

	var query *SearchQuery
	modeName, found := commandName2mode[commandName]
	if found {
		query = &SearchQuery{Mode: getMode(modeName)}
	}

	if commandName == "rule" {
		opts := i.ApplicationCommandData().Options
		if len(opts) > 0 {
			query = &SearchQuery{Mode: getMode("BANKARA"), Rule: opts[0].Value.(string)}
		}
	}

	// check command structure is valid
	if query == nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid command!",
			},
		})
		if err != nil {
			logger.Sugar().Error(err)
		}
	}
	// if valid, query to schedule store
	scheduleStore.MaybeRefresh()
	sr := scheduleStore.Search(query)

	// reply
	var err error
	if sr.Found {
		var embeds []*discordgo.MessageEmbed
		if len(sr.Slots) > 1 {
			embeds = createTwoStageInfoEmbeds(sr)
		} else {
			embeds = append(embeds, createSingleStageInfoEmbed(sr))
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: embeds,
			},
		})
	} else {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Not Found!",
			},
		})
	}
	if err != nil {
		logger.Sugar().Error(err)
	}
}
