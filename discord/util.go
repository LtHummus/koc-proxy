package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// getSourceID gets the source identifier of the message
func getSourceID(i *discordgo.InteractionCreate) string {
	if i.User != nil {
		return i.User.ID
	}

	return i.Member.User.ID
}

func isSourceAdmin(i *discordgo.InteractionCreate) bool {
	roleID := viper.GetString("discord.adminRoleID")
	if i.Member == nil {
		return false
	}

	for _, curr := range i.Member.Roles {
		if curr == roleID {
			return true
		}
	}

	return false
}

// generateEphemeralMessage generates an ephemeral message (i.e. one that only the receiving user can see) with the
// given string contents. See https://support.discord.com/hc/en-us/articles/1500000580222-Ephemeral-Messages-FAQ
func generateEphemeralMessage(msg string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: msg,
		},
	}
}

// generateChannelMessage generates a message that can be used to respond in the given channel for the interaction
// that generated it.
func generateChannelMessage(msg string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	}
}

// sendAdminMessage takes a message struct and sends it to the defined admin channel
func sendAdminMessage(d *discordgo.Session, msg *discordgo.MessageSend) {
	adminChannelID := viper.GetString("discord.adminChannelID")
	m, err := d.ChannelMessageSendComplex(adminChannelID, msg)
	if err != nil {
		log.Error().Err(err).Str("admin_channel_id", adminChannelID).Msg("could not send admin message")
	} else {
		log.Trace().Str("message_id", m.ID).Msg("admin message sent")
	}
}

func generateAdminChannelMessage(msg string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	}
}

// respondWith is a helper function that wraps an interaction handler
func respondWith(d *discordgo.Session, i *discordgo.InteractionCreate, generator func(d *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse) {
	err := d.InteractionRespond(i.Interaction, generator(d, i))
	if err != nil {
		log.Error().Stack().Err(err).Msg("error responding to interaction")
	}
}
