package handlers

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/bwmarrin/discordgo"
)

// PingCommand for getting help descriptions.
func PingCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "ping",
		Permission:      commands.All,
		HelpDescription: "Check if OutBot is online",
		Handler:         HandlePing,
		Help: commands.Help{
			Summary:             "Check if OutBot is online",
			DetailedDescription: "Check if OutBot is online.",
			Syntax:              "ping",
			Example:             "ping",
		},
	}
}

func HandlePing(s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong! v2")
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}
}
