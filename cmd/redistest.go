package cmd

import (
	"github.com/lthummus/koc-proxy/util"
	"github.com/lthummus/koc-proxy/vredis"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var redisTest = &cobra.Command{
	Use:    "redis",
	Short:  "Used for testing Redis stuff",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := vredis.Connect(cmd.Context())
		util.FatalIfError(err, "could not connect")

		count, err := vredis.GetConnectedCount(cmd.Context())
		util.FatalIfError(err, "could not get count")

		log.Info().Int64("connected_players", count).Msg("got connected count")

		names, err := vredis.GetConnectedUsernames(cmd.Context())
		util.FatalIfError(err, "could not get connected names")
		log.Info().Strs("connected_users", names).Msg("got names")

		vredis.Disconnect(cmd.Context())
	},
}
