package cmd

import (
	"github.com/lthummus/koc-proxy/util"
	"github.com/lthummus/koc-proxy/vbackend"
	"github.com/lthummus/koc-proxy/vredis"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var skipLoggedInCheck bool

func init() {
	kickCommand.Flags().BoolVarP(&skipLoggedInCheck, "skip-logged-in-check", "s", false, "don't check if the user is logged in")
}

var kickCommand = &cobra.Command{
	Use:   "kick",
	Short: "Kick a connected user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		usernameToKick := args[0]
		log.Info().Str("username", usernameToKick).Msg("kicking user")

		if !skipLoggedInCheck {
			err := vredis.Connect(cmd.Context())
			util.FatalIfError(err, "could not connect to redis")

			connected, err := vredis.IsUserConnected(cmd.Context(), usernameToKick)
			util.FatalIfError(err, "could not check if user is logged in")

			if !connected {
				log.Fatal().Str("username", usernameToKick).Msg("not connected to server")
			}
		}

		token, err := vbackend.GetAuthToken(cmd.Context(), usernameToKick)
		util.FatalIfError(err, "could not get auth token")

		err = vbackend.KickUser(cmd.Context(), token)
		util.FatalIfError(err, "could not kick user")
	},
}
