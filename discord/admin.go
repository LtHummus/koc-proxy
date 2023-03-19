package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lthummus/koc-proxy/authdb"
	"github.com/lthummus/koc-proxy/vbackend"
	"github.com/lthummus/koc-proxy/vredis"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func generateAdminResponse(d *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	options := i.ApplicationCommandData().Options

	command := options[0].Name
	switch command {
	case "whois":
		return generateAdminWhoisResponse(d, i)
	case "ban":
		return generateAdminBanMessageResponse(d, i)
	}

	return generateEphemeralMessage("something went wrong")
}

func generateAdminWhoisResponse(d *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	log.Trace().Msg("inside generateAdminWhoisResponse")
	user := i.ApplicationCommandData().Options[0].Options[0].UserValue(d)

	kocUser, err := authdb.GetByDiscordID(context.Background(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("could not query for user")
		return generateEphemeralMessage("database error")
	}

	if kocUser == nil {
		return generateEphemeralMessage("that user is not enrolled")
	}

	sendAdminMessage(d, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "User Information",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "KnockoutCity Name",
					Value: kocUser.Username,
				},
				{
					Name:  "Is Currently Banned?",
					Value: fmt.Sprintf("%v", kocUser.IsBanned()),
				},
			},
		},
	})

	return generateEphemeralMessage(fmt.Sprintf("got user %s", user.ID))
}

func generateAdminBanMessageResponse(d *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	banOptions := i.ApplicationCommandData().Options[0]

	userToBan := banOptions.Options[0].UserValue(d)

	u, err := authdb.GetByDiscordID(context.Background(), userToBan.ID)
	if err != nil {
		log.Error().Err(err).Msg("could not query for user to ban")
		return generateEphemeralMessage("database error")
	}

	if u == nil {
		log.Warn().Str("user_id", userToBan.ID).Msg("tried to ban not enrolled user")
		return generateEphemeralMessage("That user is currently not enrolled")
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("ban_modal_%s", userToBan.ID),
			Title:    fmt.Sprintf("Ban Information Form -- Banning %s", userToBan.Username),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "reason",
							Label:    "Reason for ban",
							Style:    discordgo.TextInputShort,
							Required: true,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "duration_minutes",
							Label:     "How long to ban for (in minutes)?",
							Style:     discordgo.TextInputShort,
							Required:  true,
							MinLength: 1,
							MaxLength: 5,
						},
					},
				},
			},
		},
	}
}

func handleUserBanModal(d *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ModalSubmitInteractionData) {
	log.Trace().Msg("in handleUserBanModal")

	reason := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	durationString := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	userToBan := strings.TrimPrefix(data.CustomID, "ban_modal_")

	reporterID := getSourceID(i)
	reporterMember, err := d.GuildMember(viper.GetString("discord.guildID"), reporterID)
	if err != nil {
		log.Error().Err(err).Str("member_id", reporterID).Str("guild_id", viper.GetString("discord.guildID")).Msg("could not query for banning member")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("could not get your user info"))
		return
	}

	banMember, err := d.GuildMember(viper.GetString("discord.guildID"), userToBan)
	if err != nil {
		log.Error().Err(err).Str("member_id", userToBan).Str("guild_id", viper.GetString("discord.guildID")).Msg("could not get ban target member")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("could not get ban target user info"))
		return
	}

	log.Info().
		Str("reason", reason).
		Str("duration_string", durationString).
		Str("user", userToBan).
		Str("banner", reporterMember.Nick).
		Msg("got the ban result")

	durationMinutes, err := strconv.Atoi(durationString)
	if err != nil {
		log.Error().Err(err).Msg("could not parse duration")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage(fmt.Sprintf("Could not parse %s as a duration", durationString)))
		return
	}

	u, err := authdb.GetByDiscordID(context.Background(), userToBan)
	if err != nil {
		log.Error().Err(err).Msg("could not query for user to ban")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("database error"))
		return
	}

	if u == nil {
		log.Warn().Str("user_id", userToBan).Msg("tried to ban not enrolled user")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("That user is currently not enrolled"))
		return
	}

	banExpiration := time.Now().Add(time.Duration(durationMinutes) * time.Minute)

	err = authdb.InstituteBan(context.Background(), u.Id, reason, banExpiration, reporterID)
	if err != nil {
		log.Error().Err(err).Msg("could not ban user")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("database error"))
		return
	}

	log.Info().
		Str("banned_koc_user", u.Username).
		Str("banned_discord_nick", banMember.Nick).
		Str("banned_by", reporterMember.Nick).
		Str("reason", reason).
		Time("banned_until", banExpiration).
		Msg("ban instituted")

	unixTime := banExpiration.In(time.UTC)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		connected, err := vredis.IsUserConnected(ctx, u.Username)
		if err != nil {
			log.
				Error().
				Err(err).
				Str("username", u.Username).
				Msg("could not check user connect state, pretending they are connected to force disconnect")
			connected = true
			return
		}

		if !connected {
			log.Info().Str("username", u.Username).Msg("not connected, skipping disconnection process")
			return
		}

		log.Info().Str("user", u.Username).Msg("starting user disconnect process")
		token, err := vbackend.GetAuthToken(ctx, u.Username)
		if err != nil {
			log.Error().Err(err).Str("username", u.Username).Msg("could not get auth token")
			return
		}

		err = vbackend.KickUser(ctx, token)
		if err != nil {
			log.Error().Err(err).Str("username", u.Username).Msg("could not kick user")
			return
		}

		log.Info().Str("username", u.Username).Msg("user force disconnected")
	}()

	sendAdminMessage(d, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "User banned",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "KOC User",
					Value: u.Username,
				},
				{
					Name:  "Discord Nick",
					Value: banMember.Nick,
				},
				{
					Name:  "Banned by",
					Value: reporterMember.Mention(),
				},
				{
					Name:  "Reason",
					Value: reason,
				},
				{
					Name:  "Banned Duration",
					Value: fmt.Sprintf("%d minutes (until <t:%d:f>)", durationMinutes, unixTime.Unix()),
				},
			},
		},
	})

	d.InteractionRespond(i.Interaction, generateEphemeralMessage("ok"))
}
