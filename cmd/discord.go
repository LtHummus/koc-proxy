package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/lthummus/koc-proxy/authdb"
	"github.com/lthummus/koc-proxy/discord"
	"github.com/lthummus/koc-proxy/util"
	"github.com/lthummus/koc-proxy/vredis"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bwmarrin/discordgo"
)

var discordCmd = &cobra.Command{
	Use:   "discord",
	Short: "Run the KOC Server Auth Discord bot",
	Run: func(cmd *cobra.Command, args []string) {
		if !disableDatabase {
			err := authdb.Connect(cmd.Context())
			util.FatalIfError(err, "could not connect to database")

			err = vredis.Connect(cmd.Context())
			util.FatalIfError(err, "could not connect to redis")
		}

		guildID := viper.GetString("discord.guildID")
		applicationID := viper.GetString("discord.applicationID")
		botControlChannelID := viper.GetString("discord.botChannelID")
		adminChannelID := viper.GetString("discord.adminChannelID")
		authdRoleID := viper.GetString("discord.authdRoleID")

		log.Info().
			Str("guild_id", guildID).
			Str("application_id", applicationID).
			Str("bot_control_channel_id", botControlChannelID).
			Str("admin_channel_id", adminChannelID).
			Str("authd_role_id", authdRoleID).
			Msg("running in discord mode")

		if authdRoleID == "" {
			log.Warn().Msg("authd_role_id not set; will not grant roles")
		}

		token := fmt.Sprintf("Bot %s", viper.GetString("discord.token"))
		d, err := discordgo.New(token)
		util.FatalIfError(err, "could not create discord api")

		var messageID string

		d.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
			log.Info().Str("username", d.State.User.Username).Msg("discord bot connected")

			m, err := d.ChannelMessageSendComplex(botControlChannelID, discord.InitialMessage)
			util.FatalIfError(err, "could not send open message")

			messageID = m.ID
			log.Info().Str("message_id", messageID).Msg("posted welcome message")
		})

		var cmdIds []string

		for _, curr := range discord.Commands {
			c, err := d.ApplicationCommandCreate(applicationID, guildID, curr)
			cobra.CheckErr(err)
			log.Info().Str("cmd", curr.Name).Str("id", c.ID).Msg("command created")
			cmdIds = append(cmdIds, c.ID)
		}

		d.AddHandler(func(d *discordgo.Session, i *discordgo.InteractionCreate) {
			switch i.Type {
			case discordgo.InteractionMessageComponent:
				if h, ok := discord.Handlers[i.MessageComponentData().CustomID]; ok {
					h(d, i)
				}
			case discordgo.InteractionApplicationCommand:
				if h, ok := discord.Handlers[i.ApplicationCommandData().Name]; ok {
					h(d, i)
				}
			case discordgo.InteractionModalSubmit:
				discord.HandleModalSubmit(d, i)
			}

		})

		err = d.Open()
		util.FatalIfError(err, "could not connect to discord")
		defer d.Close()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		log.Info().Msg("doing graceful shutdown")

		for _, curr := range cmdIds {
			err := d.ApplicationCommandDelete(applicationID, guildID, curr)
			util.FatalIfError(err, "could not cleanup command")
			log.Info().Str("id", curr).Msg("command deleted")
		}
		log.Info().Msg("app commands deleted")

		err = d.ChannelMessageDelete(botControlChannelID, messageID)
		util.FatalIfError(err, "could not delete channel message")
		log.Info().Str("message_id", messageID).Msg("deleted welcome message")

		if !disableDatabase {
			err = authdb.Disconnect(cmd.Context())
			util.FatalIfError(err, "could not clean up db connection")
			log.Info().Msg("db connection closed")

			vredis.Disconnect(cmd.Context())
			log.Info().Msg("redis connection closed")
		}

	},
}
