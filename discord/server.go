package discord

import (
	"context"
	"fmt"

	"github.com/lthummus/koc-proxy/vredis"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func generateServerStatusResponse(d *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	count, err := vredis.GetConnectedCount(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("could not get connected count")
		return generateEphemeralMessage("database error")
	}

	return generateChannelMessage(fmt.Sprintf("There are %d user(s) connected", count))
}
