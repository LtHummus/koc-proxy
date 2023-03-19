package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/lthummus/koc-proxy/util"
)

// KOCUser represents a user in our database. We store the SecretHash as it will be sent to the server by the KOC Client
// (i.e. FNV64a, big endian int). I'm not sure if this is a great idea, but it is what it is
type KOCUser struct {
	Id           string
	Username     string
	SecretHash   uint64
	BannedUntil  *time.Time
	DiscordID    string
	BannedReason *string
}

func (k *KOCUser) IsBanned() bool {
	if k.BannedUntil == nil {
		return false
	}

	return time.Now().Before(*k.BannedUntil)
}

// ConnectionString returns the proper command line arguments for the user to connect to our backend. If the password
// isn't in the object, a placeholder will be used instead
func (k *KOCUser) ConnectionString(password string) string {
	var usernameParam string
	if strings.Contains(k.Username, " ") {
		usernameParam = fmt.Sprintf(`-username="%s"`, k.Username)
	} else {
		usernameParam = fmt.Sprintf("-username=%s", k.Username)
	}

	var passwordParam string
	if password == "" {
		passwordParam = `-secret=<your account secret>`
	} else {
		passwordParam = fmt.Sprintf(`-secret=%s`, password)
	}

	return fmt.Sprintf("KnockoutCity.exe -backend=%s %s %s", util.GetProxyHost(), usernameParam, passwordParam)
}
