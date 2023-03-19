package auth

import (
	"context"
	"hash/fnv"

	"github.com/lthummus/koc-proxy/authdb"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// ServerSecret returns the hash of the secret that the server is expecting
func ServerSecret() uint64 {
	return GenerateHash(viper.GetString("proxy.backend.secret"))
}

// GenerateHash takes a given password and generates the fnv64a sum. Knockout City (at least the private server edition
// uses FNV64A to hash the secret for the auth payload. I'm not sure why this was chosen since it's a non-cryptographic
// hash, but could just be for some light obfuscation (since it's all otherwise plaintext)? Maybe that it's canonically
// expressible as a uint64 (big-endian)? Your guess is as good as mine ¯\_(ツ)_/¯
func GenerateHash(input string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(input)) // by spec, this never returns an error
	return h.Sum64()
}

// CheckAuth is called by our proxy to check the username + secret hash combination of the given user. This should
// probably be moved somewhere else, but it lives here for now.
func CheckAuth(ctx context.Context, username string, secret uint64) bool {
	user, err := authdb.GetByUsername(ctx, username)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("could not query for user")
		return false
	}
	if user == nil {
		log.Warn().Str("username", username).Msg("login attempt no username")
		return false
	}
	if user.SecretHash != secret {
		log.Warn().Str("username", username).Msg("incorrect password")
		return false
	}

	return true

}
