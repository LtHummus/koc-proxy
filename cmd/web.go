package cmd

import (
	"github.com/lthummus/koc-proxy/util"
	"github.com/lthummus/koc-proxy/vredis"
	"github.com/lthummus/koc-proxy/webserver"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Run the status webserver",
	Run: func(cmd *cobra.Command, args []string) {
		if !disableDatabase {
			err := vredis.Connect(cmd.Context())
			util.FatalIfError(err, "could not connect to redis")
		}

		if viper.GetBool("web.tls.enabled") {
			webserver.RunHTTPSServer()
		} else {
			webserver.RunHTTPServer()
		}
	},
}
