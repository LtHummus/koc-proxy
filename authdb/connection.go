package authdb

import (
	"fmt"

	"github.com/spf13/viper"
)

func buildConnectionString(scheme string) string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		scheme,
		viper.GetString("auth.db.username"),
		viper.GetString("auth.db.password"),
		viper.GetString("auth.db.host"),
		viper.GetInt32("auth.db.port"),
		viper.GetString("auth.db.name"))
}
