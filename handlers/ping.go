package handlers

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/bwmarrin/discordgo"
)

func HandlePing(s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong! v2")
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}
}
