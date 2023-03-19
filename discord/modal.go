package discord

import (
	"context"
	"fmt"

	"github.com/lthummus/koc-proxy/auth"
	"github.com/lthummus/koc-proxy/authdb"
	"github.com/lthummus/koc-proxy/types"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func handleUserEnrollModal(d *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ModalSubmitInteractionData) {
	log.Trace().Msg("in handleUserEnrollModal")

	username := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	log.Info().Str("username", username).Msg("got request to create user")

	isEnrolled, err := authdb.IsAlreadyEnrolled(context.Background(), getSourceID(i))
	if err != nil {
		log.Error().Err(err).Msg("could not check enrollment state")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("Database error. Contact LtHummus about this"))
		return
	}

	if isEnrolled {
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("It looks like you've already got an account. Use the reset secret command if you forgot it"))
		return
	}

	// TODO: this is a case sensitive check, we should probably do some INDEX LOWER(username) or whatever so we can
	//       quickly check for usernames case-insensitively
	existingUser, err := authdb.GetByUsername(context.Background(), username)
	if err != nil {
		log.Error().Err(err).Msg("could not check for username that exists")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("Database error. Contact LtHummus about this"))
		return
	}

	if existingUser != nil {
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("That username has been taken. Try running `/auth enroll` again"))
		return
	}

	if roleID := viper.GetString("discord.authdRoleID"); roleID != "" {
		err = d.GuildMemberRoleAdd(viper.GetString("discord.guildID"), getSourceID(i), roleID)
		if err != nil {
			log.Error().Err(err).Str("authd_role_id", roleID).Msg("could not grant authd role")
			d.InteractionRespond(i.Interaction, generateEphemeralMessage("could not grant you the auth'd role"))
			return
		}
	}

	password := types.GeneratePassword()

	u := &types.KOCUser{
		Username:   username,
		DiscordID:  getSourceID(i),
		SecretHash: auth.GenerateHash(password),
	}

	err = authdb.CreateUser(context.Background(), u)
	if err != nil {
		log.Error().Err(err).Msg("error creating user")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("Unable to persist user account. Contact LtHummus about this"))
		return
	}

	var response = fmt.Sprintf("You are all set. Launch the game with `%s`", u.ConnectionString(password))

	d.InteractionRespond(i.Interaction, generateEphemeralMessage(response))
}
