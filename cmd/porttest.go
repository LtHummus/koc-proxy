package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"net/http"
)

var portTestCommand = &cobra.Command{
	Use:    "porttest",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Int("port", 5434).Msg("listening on port")
		http.ListenAndServe(":5434", nil)
	},
}
