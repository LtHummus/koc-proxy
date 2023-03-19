package cmd

import (
	"crypto/subtle"
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/lthummus/koc-proxy/auth"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Generate a hash for a given password",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Enter password: ")
		password1, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Print("\n")

		if err != nil {
			log.Fatal().Err(err).Msg("could not read password")
		}

		fmt.Print("Enter password (this time with feeling): ")
		password2, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Print("\n")
		if err != nil {
			log.Fatal().Err(err).Msg("could not read password 2")
		}

		if subtle.ConstantTimeCompare(password1, password2) != 1 {
			log.Fatal().Msg("passwords do not match")
		}

		fmt.Printf("%d\n", auth.GenerateHash(string(password1)))
	},
}
