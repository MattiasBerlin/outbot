package handlers

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/bwmarrin/discordgo"
)

// StatusCommand for setting OutBot's status.
func StatusCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "status",
		Permission:      commands.Officers,
		HelpDescription: "Set OutBot's status",
		Handler:         HandleStatus,
		Help: commands.Help{
			Summary:             "Set OutBot's status",
			DetailedDescription: "Set status of OutBot in discord.",
			Syntax:              "status <status>",
			Example:             "status Hades Star",
		},
	}
}

func HandleStatus(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	err := s.UpdateStatus(0, msg)
	if err != nil {
		fmt.Println("Failed to update status:", err)
		return
	}

	resp := discordgo.MessageEmbed{
		Title: "Status set!",
		Color: successColor,
	}
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &resp)
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}
}
