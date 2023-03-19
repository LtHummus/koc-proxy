package discord

import (
	"context"
	"fmt"

	"github.com/lthummus/koc-proxy/auth"
	"github.com/lthummus/koc-proxy/authdb"
	"github.com/lthummus/koc-proxy/types"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func generateEnrollResponse(d *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	isEnrolled, err := authdb.IsAlreadyEnrolled(context.Background(), getSourceID(i))
	if err != nil {
		log.Error().Err(err).Msg("could not check enrollment state")
		return generateEphemeralMessage("Database error")
	}
	if isEnrolled {
		return generateEphemeralMessage("It looks like you've already enrolled. If you forgot your account secret, use the command to reset your password")
	}
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("enroll_modal_response_%s", getSourceID(i)),
			Title:    "The City Never Sleeps Enrollment",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "username",
							Label:     "Knockout City Username",
							Style:     discordgo.TextInputShort,
							Required:  true,
							MinLength: 3,
							MaxLength: 30,
						},
					},
				},
			},
		},
	}
}

func generateStatusResponse(d *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	user, err := authdb.GetByDiscordID(context.Background(), getSourceID(i))
	if err != nil {
		log.Error().Err(err).Msg("could not query database")
		return generateEphemeralMessage("Database error")
	}

	if user == nil {
		return generateEphemeralMessage("It looks like you haven't enrolled yet.")
	}

	var banMessage string
	if user.IsBanned() {
		banMessage = fmt.Sprintf("You are banned until <t:%d:f> because **%s**", user.BannedUntil.Unix(), *user.BannedReason)
	} else {
		banMessage = ""
	}

	responseText := fmt.Sprintf("You are enrolled. Your username is `%s`. Start the game with `%s`. %s", user.Username, user.ConnectionString(""), banMessage)
	return generateEphemeralMessage(responseText)
}

func generateResetSecretResponse(d *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	user, err := authdb.GetByDiscordID(context.Background(), getSourceID(i))
	if err != nil {
		log.Error().Err(err).Msg("could not query for user")
		return generateEphemeralMessage("Database error")
	}

	if user == nil {
		return generateEphemeralMessage("It looks like you're not enrolled yet")
	}

	password := types.GeneratePassword()
	hash := auth.GenerateHash(password)

	err = authdb.UpdatePassword(context.Background(), user.Id, hash)
	if err != nil {
		log.Error().Err(err).Msg("could not update user row")
		return generateEphemeralMessage("Database error")
	}

	msg := fmt.Sprintf("Your old secret has been updated and a new one has been generated. Launch with `%s`", user.ConnectionString(password))
	return generateEphemeralMessage(msg)
}
