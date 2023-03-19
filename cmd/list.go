package cmd

import (
	"github.com/lthummus/koc-proxy/util"
	"github.com/lthummus/koc-proxy/vredis"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "connected",
	Short: "List connected users",
	Run: func(cmd *cobra.Command, args []string) {
		err := vredis.Connect(cmd.Context())
		util.FatalIfError(err, "could not connect")

		connected, err := vredis.GetConnectedUsernames(cmd.Context())
		util.FatalIfError(err, "could not check connected users")

		log.Info().Strs("connected", connected).Msg("got connected users")

		vredis.Disconnect(cmd.Context())
	},
}
