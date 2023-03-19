package vredis

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var client *redis.Client

// sessionSetLock controls access to the function that queries for the name of the set that holds the current
// session ids...there _HAS_ to be a better way to do this whole user set thing, but this is what it is for now
// ...we should probably revisit it later. This is also left over from my initial implementation and we probably don't
// actually need this lock anymore...
var sessionSetLock = &sync.Mutex{}

// Connect connects to the KOC redis backend. Note that by default, the KOC redis backend has no authentication and
// is only reachable on the same machine.
func Connect(ctx context.Context) error {
	client = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("koc.redis.host"),
		Password: "",
		DB:       0,
	})

	return nil
}

// updateUserSessionSetKey inspects the keys for one that looks like `backend:users:NNN` where NNN is some unpredictable
// number (I think it increments? But who knows). This is all very loosely locked with a mutex, but there is definitely
// a better way to do this.
func updateUserSessionsSetKey(ctx context.Context) (string, error) {
	sessionSetLock.Lock()
	defer sessionSetLock.Unlock()

	userSetSearchResults := client.Keys(ctx, "backend:users:*")
	results, err := userSetSearchResults.Result()
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		log.Warn().Msg("no users connected")
		return "", nil
	}

	if len(results) != 1 {
		return "", fmt.Errorf("more than one potential backend running")
	}

	log.Info().Str("user_set_name", results[0]).Msg("found user set")
	return results[0], nil
}

// GetConnectedCount returns the number of users currently connected to the server
func GetConnectedCount(ctx context.Context) (int64, error) {
	key, err := updateUserSessionsSetKey(ctx)
	if err != nil {
		return 0, err
	}

	if key == "" {
		return 0, nil
	}

	res := client.SCard(ctx, key)
	return res.Result()
}

// getUserSessionIDs returns a list of the ids that it finds in the given `backend:users:nnn` set. As far as I can tell
// the way the backend keeps track of these sort of things is that it keeps a list of these ephemeral ids in here and then
// there is a hash table named `user:session:NNN` where N is the id number that has the connected user info.
func getUserSessionIDs(ctx context.Context) ([]string, error) {
	key, err := updateUserSessionsSetKey(ctx)
	if err != nil {
		return nil, err
	}

	if key == "" {
		return nil, nil
	}

	userDocumentKeys, err := client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return userDocumentKeys, nil
}

func getUsernameFromDoc(ctx context.Context, docID string) (string, error) {
	document, err := client.HGetAll(ctx, fmt.Sprintf("user:session:%s", docID)).Result()
	if err != nil {
		log.Error().Err(err).Str("document_name", docID).Msg("could not retrieve document")
		return "", err
	}

	return document["username"], nil
}

func IsUserConnected(ctx context.Context, needle string) (bool, error) {
	userDocumentKeys, err := getUserSessionIDs(ctx)
	if err != nil {
		return false, err
	}

	for _, curr := range userDocumentKeys {
		username, err := getUsernameFromDoc(ctx, curr)
		if err != nil {
			log.Error().Err(err).Str("document_id", curr).Msg("could not retrieve document")
			return false, err
		}

		if username == needle {
			return true, nil
		}
	}

	return false, nil
}

func GetConnectedUsernames(ctx context.Context) ([]string, error) {
	userDocumentKeys, err := getUserSessionIDs(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(userDocumentKeys))
	for i, curr := range userDocumentKeys {
		username, err := getUsernameFromDoc(ctx, curr)
		if err != nil {
			log.Error().Err(err).Str("document_id", curr).Msg("could not retrieve document")
			return nil, err
		}
		names[i] = username
	}

	return names, nil
}

func Disconnect(ctx context.Context) {
	client.Close()
}
