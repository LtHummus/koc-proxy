package authdb

import (
	"context"
	"time"

	"github.com/lthummus/koc-proxy/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// TODO: this whole thing should probably be refactored so this lives inside a struct instead of being a global
var conn *pgxpool.Pool

// Connect opens a connection to the PostgreSQL database used for storing our data
func Connect(ctx context.Context) error {
	config, err := pgxpool.ParseConfig(buildConnectionString("postgres"))
	if err != nil {
		return err
	}

	conn, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	log.Info().Str("host", config.ConnConfig.Host).Uint16("port", config.ConnConfig.Port).Str("username", config.ConnConfig.User).Msg("auth connected")

	err = conn.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}

func Disconnect(ctx context.Context) error {
	if conn == nil {
		log.Warn().Msg("attempted to disconnect from a db that was never connected")
		return nil
	}
	conn.Close()
	return nil
}

func CreateUser(ctx context.Context, input *types.KOCUser) error {
	uid := uuid.New().String()
	res, err := conn.Exec(ctx,
		"INSERT INTO users (id, username, secret_hash, discord_id) VALUES ($1, $2, $3, $4)",
		uid, input.Username, input.SecretHash, input.DiscordID)
	if err != nil {
		return err
	}

	log.Info().Int64("rows_affected", res.RowsAffected()).Msg("user created")
	return nil
}

func IsAlreadyEnrolled(ctx context.Context, discordID string) (bool, error) {
	// TODO: can this be cleaned up
	log.Debug().Str("discord_id", discordID).Msg("checking to see if user is enrolled")
	var id string
	err := conn.QueryRow(ctx, "SELECT id FROM users WHERE discord_id = $1", discordID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return true, err
	}

	return true, nil
}

// GetByUsername gets a user object from the database that matches the given username. If no user is found, no error will
// be returned and will return a nil pointer
func GetByUsername(ctx context.Context, username string) (*types.KOCUser, error) {
	var id string
	var hash uint64
	var bannedUntil *time.Time
	var discordID string
	var banReason *string
	err := conn.QueryRow(ctx, "SELECT id, secret_hash, banned_until, discord_id, banned_reason FROM users WHERE username = $1", username).Scan(&id, &hash, &bannedUntil, &discordID, &banReason)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &types.KOCUser{
		Id:           id,
		Username:     username,
		SecretHash:   hash,
		BannedUntil:  bannedUntil,
		DiscordID:    discordID,
		BannedReason: banReason,
	}, nil
}

func UpdatePassword(ctx context.Context, id string, secretHash uint64) error {
	log.Debug().Str("id", id).Msg("updating password")
	res, err := conn.Exec(ctx, "UPDATE users SET secret_hash = $1 WHERE id = $2", secretHash, id)
	if err != nil {
		return err
	}
	log.Debug().Int64("rows_affected", res.RowsAffected()).Msg("updated")
	return nil
}

func GetByDiscordID(ctx context.Context, discordID string) (*types.KOCUser, error) {
	log.Debug().Str("discord_id", discordID).Msg("querying for user")
	var id string
	var username string
	var hash uint64
	var bannedUntil *time.Time
	var banReason *string
	err := conn.QueryRow(ctx, "SELECT id, username, secret_hash, banned_until, banned_reason FROM users WHERE discord_id = $1", discordID).Scan(&id, &username, &hash, &bannedUntil, &banReason)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &types.KOCUser{
		Id:           id,
		Username:     username,
		SecretHash:   hash,
		BannedUntil:  bannedUntil,
		DiscordID:    discordID,
		BannedReason: banReason,
	}, nil
}

func InstituteBan(ctx context.Context, userGuid string, reason string, expiration time.Time, bannerDiscordID string) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "INSERT INTO ban_records (id, user_guid, ban_time, ban_expiration, reporter_discord_id, ban_reason) VALUES ($1, $2, $3, $4, $5, $6)",
		"", userGuid, time.Now(), expiration, bannerDiscordID, reason)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE users SET banned_until = $1, ban_reason = $2 WHERE id = $3", expiration, reason, userGuid)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
