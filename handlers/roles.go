package handlers

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/bwmarrin/discordgo"
)

type Role string

const (
	currentWhitestar Role = "442643047541374977"
)

// ClearParticipantsCommand for clearing the participation list.
func SetRolesCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "setroles",
		Permission:      commands.Officers,
		HelpDescription: "Set roles after a WS match is found",
		Handler:         HandleSetRoles,
		Help: commands.Help{
			Summary:             "Set roles after a WS match is found",
			DetailedDescription: "Set roles after a WS match is found.",
			Syntax:              "setroles [instance]",
			Example:             "setroles B",
		},
	}
}

func setRole(s *discordgo.Session, guildID, userID string, role Role) {
	err := s.GuildMemberRoleAdd(guildID, userID, string(role))
	if err != nil {
		fmt.Printf("Failed to set role %v for user %v: %v\n", role, userID, err)
		return
	}
}

func removeRole(s *discordgo.Session, guildID, userID string, role Role) {
	err := s.GuildMemberRoleRemove(guildID, userID, string(role))
	if err != nil {
		fmt.Printf("Failed to remove role %v for user %v: %v\n", role, userID, err)
		return
	}
}

// removeRolesForParticipants and return the amount of users affected.
func removeRolesForParticipants(instance instance, s *discordgo.Session, db *sql.DB, guildID string) int {
	participants, err := getParticipantsFromDatabase(db, instance)
	if err != nil {
		fmt.Println("Failed to get participants:", err.Error())
		return 0
	}

	for _, p := range participants {
		go removeRole(s, guildID, p.userID, currentWhitestar)
	}

	return len(participants)
}

func HandleSetRoles(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	participants, err := getParticipantsFromDatabase(db, channelToInstance(m.ChannelID, msg))
	if err != nil {
		fmt.Println("Failed to get participants:", err.Error())
		return
	}

	var participating int
	for _, p := range participants {
		if p.participating {
			participating++
			setRole(s, guildID, p.userID, currentWhitestar)
		}
	}

	response := discordgo.MessageEmbed{
		Color:       successColor,
		Description: fmt.Sprintf("Set Current Whitestar role for %d members!", participating),
	}
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &response)
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}
