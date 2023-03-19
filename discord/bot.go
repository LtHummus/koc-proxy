package discord

import (
	"strings"

	"github.com/lthummus/koc-proxy/util"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// InitialMessage is the opening message we send to the discord user-facing channel. This message has buttons that
// users can use to interact with the bot, including signing up, rotating their secret, and getting their user statsu
var InitialMessage = &discordgo.MessageSend{
	Content: "The City Never Sleeps KO City Auth bot. This bot will generate a key for you in order to set up your secret key so you can play on the server. Click the sign up button below to get started!",
	Components: []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Signup for an account",
					Style:    discordgo.PrimaryButton,
					Disabled: false,
					CustomID: "enroll_button",
				},
				discordgo.Button{
					Label:    "User status",
					Style:    discordgo.SecondaryButton,
					Disabled: false,
					CustomID: "user_status_button",
				},
				discordgo.Button{
					Label:    "Change Server Password",
					Style:    discordgo.DangerButton,
					Disabled: false,
					CustomID: "change_password_button",
				},
			},
		},
	},
}

// Commands is a slice of *discordgo.ApplicationCommand structs (that each represent a slash-command) that we populate
// the discord server with. See https://discord.com/developers/docs/interactions/application-commands for more details
var Commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "admin",
		Description:              "Do an admin thing",
		DMPermission:             util.P(false),
		DefaultMemberPermissions: util.P[int64](discordgo.PermissionModerateMembers),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "whois",
				Description: "get KOC info about a user",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "user",
						Description: "user to query",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
			{
				Name:        "ban",
				Description: "ban a user",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "user",
						Description: "user to ban",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
		},
	},
	{
		Name:        "status",
		Description: "Get Server Status",
	},
}

// Handlers holds a map of strings (representing actions) to functions to handle those interactions. The functions
// must have a signature of func(d *discordgo.Session, i *discordgo.InteractionCreate). Ideally, these functions
// will respond with some sort of interaction in their course of doing business.
var Handlers = map[string]func(d *discordgo.Session, i *discordgo.InteractionCreate){
	"enroll_button": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Trace().Msg("enroll_button proxy")
		respondWith(d, i, generateEnrollResponse)
	},
	"user_status_button": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Trace().Msg("user_status_button proxy")
		respondWith(d, i, generateStatusResponse)
	},
	"change_password_button": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Trace().Msg("change_password_button")
		respondWith(d, i, generateResetSecretResponse)
	},
	"status": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Trace().Msg("status")
		respondWith(d, i, generateServerStatusResponse)
	},
	"admin": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Trace().Msg("admin")
		respondWith(d, i, generateAdminResponse)
	},
}

// HandleModalSubmit is a special function that handles modal submissions. There are multiple modals in use in the
// application, so this finds which kind the user has responded to and dispatches the message to the proper handler in
// modal.go
func HandleModalSubmit(d *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	if strings.HasPrefix(data.CustomID, "enroll_modal_response_") {
		handleUserEnrollModal(d, i, data)
	} else if strings.HasPrefix(data.CustomID, "ban_modal_") {
		handleUserBanModal(d, i, data)
	} else {
		log.Warn().Str("custom_id", data.CustomID).Msg("unknown custom id in modal submit")
		d.InteractionRespond(i.Interaction, generateEphemeralMessage("Unknown modal submission? Contact LtHummus if this happens"))
		return
	}
}
