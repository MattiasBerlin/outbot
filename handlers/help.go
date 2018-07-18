package handlers

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/bwmarrin/discordgo"
)

func HandleHelp(s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	helpMsg := "**Command list:**"

	for _, cmd := range cmds {
		description := cmd.HelpDescription
		if description == "" {
			description = "*No description available (yell at Maro)*"
		}

		helpMsg += fmt.Sprintf("\n%v - %v", cmd.CallPhrase, description)
	}

	_, err := s.ChannelMessageSend(m.ChannelID, helpMsg)
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}
}
