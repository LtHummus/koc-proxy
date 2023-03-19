package vbackend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lthummus/koc-proxy/auth"
	"github.com/lthummus/koc-proxy/types"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// KickUser forces a connected user to disconnect from the KOC backend.  We don't really have a direct way of accomplishing
// this, so this function takes the user's auth token (which can be grabbed with the GetAuthToken function) and we open up
// the websocket as that user. That causes the currently connected user to disconnect immediately with a "duplicate login"
// error message on their client. We then immediately close the websocket channel.
//
// One thing that might be cleaner is to build this in to the proxy itself and track which connections belong to which
// players (this will potentially involve proxying the websocket traffic as well (or maybe just the upgrade request?)
// and inspecting auth tokens to remember which connections involve which players. For now, though, we do it this way.
func KickUser(ctx context.Context, authToken string) error {
	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	headers.Set("Game-Server-IP", viper.GetString("proxy.backend.upstream"))
	headers.Set("Game-Executable-Version", GameVersionString)
	headers.Set("Game-Version", GameVersionString)
	headers.Set("Game-Reconnect", "0")

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, fmt.Sprintf("ws://%s", viper.GetString("proxy.backend.upstream")), headers)
	if err != nil {
		return err
	}

	// we don't actually have to send any data or do anything with the websocket connection, just connecting as the
	// target user is enough, so we can just close the connection immediately and call it a day
	return conn.Close()
}

// GetAuthToken grabs an auth token for the given username from the backend.
func GetAuthToken(ctx context.Context, username string) (string, error) {
	systemGuid, err := uuid.NewV1() // uses v1 because i guess that's mac address based + date/time?
	if err != nil {
		return "", err
	}
	authPayload := &types.AuthPayload{
		AuthProvider: "dev",
		Credentials: &types.AuthCredentials{
			Username:            username,
			Secret:              auth.ServerSecret(),
			Platform:            Platform,
			SystemGuid:          systemGuid.String(),
			Version:             GameVersion,
			Build:               GameBuild,
			BootSessionGuid:     BootSessionID,
			IsUsingEpicLauncher: false,
			PID:                 PID,
		},
	}

	serializedPayload, err := json.Marshal(authPayload)
	if err != nil {
		return "", err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/api/auth", viper.GetString("proxy.backend.upstream")), bytes.NewBuffer(serializedPayload))
	if err != nil {
		return "", err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Content-Length", fmt.Sprintf("%d", len(serializedPayload)))

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		var authError *types.AuthResponseError
		err = json.NewDecoder(res.Body).Decode(&authError)
		if err != nil {
			log.Error().Err(err).Msg("error when decoding errored payload from auth")
			return "", err
		}
		log.Error().Str("message", authError.Message).Msg("could not get auth token")
		return "", fmt.Errorf("could not get auth token")
	}

	var response *types.AuthResponseSuccess
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.Token, nil
}
