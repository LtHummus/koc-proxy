package types

import (
	"encoding/json"
	"math/rand" // TODO: replace with crypt/rand at some point

	"github.com/rs/zerolog/log"
)

var (
	IncorrectPasswordResponse       []byte
	IncorrectPasswordResponseLength int
)

func init() {
	// this is what the KOC backend server sends back if the secret is incorrect
	p, err := json.Marshal(AuthResponseError{
		Error:   "Unprocessable Entity",
		Message: "incorrect -secret=", // yes, it is -secret= with a dash and an equals sign...no idea
	})
	if err != nil {
		log.Fatal().Err(err).Msg("could not encode secret")
	}
	IncorrectPasswordResponse = p
	IncorrectPasswordResponseLength = len(p)
}

type AuthCredentials struct {
	Username            string `json:"username"`
	Secret              uint64 `json:"secret"`
	Platform            string `json:"platform"`
	SystemGuid          string `json:"system_guid"`
	Version             int    `json:"version"`
	Build               string `json:"build"`
	BootSessionGuid     string `json:"boot_session_guid"`
	IsUsingEpicLauncher bool   `json:"is_using_epic_launcher"`
	PID                 uint64 `json:"pid"`
}

type AuthPayload struct {
	Credentials  *AuthCredentials `json:"credentials"`
	AuthProvider string           `json:"auth_provider"`
}

type AuthResponseError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type AuthResponseSuccess struct {
	// there are plenty more fields in this response, but we don't care about them (for now?) so we just ignore them
	Token string `json:"token"`
}

var letterRunes = []byte("abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ0123456789")

func GeneratePassword() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
