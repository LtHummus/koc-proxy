package util

import (
	"fmt"

	"github.com/spf13/viper"
)

func GetProxyHost() string {
	return fmt.Sprintf("%s:%d", viper.GetString("proxy.host"), viper.GetInt("proxy.port"))
}
