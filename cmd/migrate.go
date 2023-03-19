package cmd

import (
	"github.com/lthummus/koc-proxy/authdb"
	"github.com/lthummus/koc-proxy/util"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var version uint

func init() {
	migrateCommand.Flags().UintVar(&version, "version", 0, "version to migrate to")
}

var migrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations to upgrade the auth database",
	Long: `Runs database migration to upgrade (or downgrade) the database to a given version. Note that if no version is
specified (with -v), we will print out the version of the database instead`,
	Run: func(cmd *cobra.Command, args []string) {
		if version == 0 {
			log.Info().Msg("getting database version information")
			err := authdb.GetVersionInfo()
			util.FatalIfError(err, "could not get version info")
			return
		}

		log.Info().Uint("version", version).Msg("starting migration")
		err := authdb.MigrateDatabase(version)
		if err != nil {
			log.Fatal().Err(err).Uint("version", version).Msg("could not complete migration")
			return
		}
		log.Info().Uint("version", version).Msg("migration completed")
	},
}
