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
	successColor = 0x00ff00
	failColor    = 0xff0000

	defaultPreferredRole = "No preference"
)

type wsRole string

const (
	defaultRole wsRole = "No preference"
	defense     wsRole = "Defense"
	offense     wsRole = "Offense"
	hunter      wsRole = "Hunter"
)

// wsRoleFromString returns a wsRole from a string.
// The parameter is case insensitive.
func wsRoleFromString(text string) wsRole {
	text = strings.ToLower(text)

	var role wsRole
	switch text {
	case "defense", "defensive", "def":
		role = defense
	case "offense", "offensive", "off":
		role = offense
	case "hunter":
		role = hunter
	default:
		role = defaultRole
	}
	return role
}

type participant struct {
	name          string
	participating bool
	preferredRole wsRole
}

// OptInCommand for opting in to white stars.
func OptInCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "optin",
		Permission:      commands.Members,
		HelpDescription: "Opt in for the next WS",
		Handler:         HandleOptIn,
		Help: commands.Help{
			Summary: "Opt in for the next WS",
			DetailedDescription: `Out in for the next White Star match.
				Accepted preferred roles: def, off, hunter`,
			Syntax:  "optin [preferred role]",
			Example: "optin defense",
		},
	}
}

// OptOutCommand for opting out of white stars.
func OptOutCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "optout",
		Permission:      commands.Members,
		HelpDescription: "Opt out for the next WS",
		Handler:         HandleOptOut,
		Help: commands.Help{
			Summary:             "Opt out of the next WS",
			DetailedDescription: "Out out of the next White Star match.",
			Syntax:              "optout",
			Example:             "optout",
		},
	}
}

// ListParticipantsCommand for listing members interest in joining the next WS.
func ListParticipantsCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "list",
		Permission:      commands.Members,
		HelpDescription: "List members interest in joining the next WS",
		Handler:         HandleListParticipants,
		Help: commands.Help{
			Summary:             "List members interest in joining the next WS",
			DetailedDescription: "List every member that has opted in respectively out from the next White Star.",
			Syntax:              "list",
			Example:             "list",
		},
	}
}

// ClearParticipantsCommand for clearing the participation list.
func ClearParticipantsCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "clear",
		Permission:      commands.Officers,
		HelpDescription: "Clear the participation list",
		Handler:         HandleClearParticipants,
		Help: commands.Help{
			Summary:             "Clear the participation list",
			DetailedDescription: "Clear the participation list.",
			Syntax:              "clear",
			Example:             "clear",
		},
	}
}

// HandleOptIn handles opt in commands.
func HandleOptIn(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	setParticipation(true, wsRoleFromString(msg), fmt.Sprintf("You've opted in, %v!", m.Author.Username), s, m, db)
}

// HandleOptOut handles opt out commands.
func HandleOptOut(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	setParticipation(false, wsRoleFromString(msg), fmt.Sprintf("You've opted out, %v!", m.Author.Username), s, m, db)
}

// HandleClearParticipants handles clearing the participation list.
func HandleClearParticipants(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	err := clearParticipantsFromDatabase(db)
	if err != nil {
		_, err = s.ChannelMessageSend(m.ChannelID, "Failed to clear participants")
		if err != nil {
			fmt.Println("Failed to send message:", err.Error())
			return
		}
		return
	}

	response := discordgo.MessageEmbed{
		Color:       successColor,
		Description: "Participation list cleared!",
	}
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &response)
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}

// HandleListParticipants handles the command for listing participants.
func HandleListParticipants(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	status, err := optStatus(db)
	if err != nil {
		fmt.Println("Failed to get participation status:", err.Error())
		status = "[Failed to get participation status]"
	}

	response := discordgo.MessageEmbed{
		Color:       successColor,
		Description: status,
	}
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &response)
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}

func setParticipation(participating bool, preferredRole wsRole, updateMessage string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB) {
	participant := participant{
		name:          m.Author.Username,
		participating: participating,
		preferredRole: preferredRole,
	}
	err := setParticipatingInDatabase(db, participant)
	if err != nil {
		fmt.Println("Failed to set participation:", err.Error())
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to set participation: %v", err))
		if err != nil {
			fmt.Println("Failed to send message:", err.Error())
			return
		}
		return
	}

	status, err := optStatus(db)
	if err != nil {
		fmt.Println("Failed to get participation status:", err.Error())
		status = "[Failed to get participation status]"
	}

	msg := discordgo.MessageEmbed{
		Color:       successColor,
		Description: fmt.Sprintf("%v\n\n%v", updateMessage, status),
	}
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &msg)
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}

func optStatus(db *sql.DB) (string, error) {
	participants, err := getParticipantsFromDatabase(db)
	if err != nil {
		return "", err
	}

	roleMap := make(map[string][]string)
	var optIn, optOut []string
	for _, p := range participants {
		if p.participating {
			optIn = append(optIn, p.name)

			roleList, exists := roleMap[string(p.preferredRole)]
			if !exists {
				roleList = []string{p.name}
			} else {
				roleList = append(roleList, p.name)
			}
			roleMap[string(p.preferredRole)] = roleList
		} else {
			optOut = append(optOut, p.name)
		}
	}

	var roles string
	for role, names := range roleMap {
		roles += fmt.Sprintf("*%v* (%v): %v\n", role, len(names), strings.Join(names, ", "))
	}

	return fmt.Sprintf("**Participants**:\n**Opted in** (%v):\n%v**Opted out** (%v):\n%v", len(optIn), roles, len(optOut), strings.Join(optOut, ", ")), nil
}

func setParticipatingInDatabase(db *sql.DB, participant participant) error {
	statement := "INSERT INTO participants (name, participating, preferred_role) VALUES ($1, $2, $3) ON CONFLICT (name) DO UPDATE SET participating = $2, preferred_role = $3"
	_, err := db.Exec(statement, participant.name, participant.participating, participant.preferredRole)
	return err
}

func getParticipantsFromDatabase(db *sql.DB) ([]participant, error) {
	rows, err := db.Query("SELECT name, participating, preferred_role FROM participants")
	if err != nil {
		return nil, errors.Wrap(err, "failed to do query")
	}
	defer rows.Close()

	var participants []participant
	for rows.Next() {
		var p participant
		err = rows.Scan(&p.name, &p.participating, &p.preferredRole)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		participants = append(participants, p)
	}

	return participants, nil
}

func clearParticipantsFromDatabase(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE TABLE participants")
	return err
}
