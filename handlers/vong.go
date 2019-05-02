package handlers

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"strings"
)

const (
	infoColor = 0x4286f4
)

// HelpCommand for getting help descriptions.
func HelpCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "vong",
		Permission:      commands.All,
		HelpDescription: "Vongs testing command",
		Handler:         HandleHelp,
		Help: commands.Help{
			Summary:             "Vongs testing command",
			DetailedDescription: "Vongs testing command",
			Syntax:              "vong [command]",
			Example:             "vong rocks",
		},
	}
}

func HandleHelp(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	splitMsg := strings.Split(msg, " ")
	_, err := s.ChannelMessageSend(m.ChannelID, "Vong always rocks!")
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}
}