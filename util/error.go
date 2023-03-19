package util

import "github.com/rs/zerolog/log"

// FatalIfError takes an error and a message. If the given error is nil, nothing happens. If the error is non-nil, it is
// fatally logged to the terminal with the given message and program execution stops
func FatalIfError(err error, message string) {
	if err != nil {
		log.Fatal().Err(err).Msg(message)
	}
}
