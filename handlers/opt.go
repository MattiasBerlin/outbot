package handlers

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"strings"
)

type participant struct {
	name          string
	participating bool
}

// OptInCommand for opting in to white stars.
func OptInCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "optin",
		Permission:      commands.Members,
		HelpDescription: "Opt in for the next WS",
		Handler:         HandleOptIn,
	}
}

// OptOutCommand for opting out of white stars.
func OptOutCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "optout",
		Permission:      commands.Members,
		HelpDescription: "Opt out for the next WS",
		Handler:         HandleOptOut,
	}
}

// ListParticipantsCommand for listing members interest in joining the next WS.
func ListParticipantsCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "list",
		Permission:      commands.Members,
		HelpDescription: "List members interest in joining the next WS",
		Handler:         HandleListParticipants,
	}
}

// ClearParticipantsCommand for clearing the participation list.
func ClearParticipantsCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "clear",
		Permission:      commands.Officers,
		HelpDescription: "Clear the participation list",
		Handler:         HandleClearParticipants,
	}
}

// HandleOptIn handles opt in commands.
func HandleOptIn(s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	setParticipation(true, fmt.Sprintf("You've opted in, %v!", m.Author.Username), s, m, db)
}

// HandleOptOut handles opt out commands.
func HandleOptOut(s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	setParticipation(false, fmt.Sprintf("You've opted out, %v!", m.Author.Username), s, m, db)
}

// HandleClearParticipants handles clearing the participation list.
func HandleClearParticipants(s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	err := clearParticipantsFromDatabase(db)
	if err != nil {
		_, err = s.ChannelMessageSend(m.ChannelID, "Failed to clear participants")
		if err != nil {
			fmt.Println("Failed to send message:", err.Error())
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Participation list cleared!")
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}

// HandleListParticipants handles the command for listing participants.
func HandleListParticipants(s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	status, err := optStatus(db)
	if err != nil {
		fmt.Println("Failed to get participation status:", err.Error())
		status = "[Failed to get participation status]"
	}

	_, err = s.ChannelMessageSend(m.ChannelID, status)
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}

func setParticipation(participating bool, updateMessage string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB) {
	participant := participant{
		name:          m.Author.Username,
		participating: participating,
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

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v\n\n%v", updateMessage, status))
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

	var optIn, optOut []string
	for _, p := range participants {
		if p.participating {
			optIn = append(optIn, p.name)
		} else {
			optOut = append(optOut, p.name)
		}
	}

	return fmt.Sprintf("**Participants**:\nOpted in (%v): %v\nOpted out (%v): %v", len(optIn), strings.Join(optIn, ", "), len(optOut), strings.Join(optOut, ", ")), nil
}

func setParticipatingInDatabase(db *sql.DB, participant participant) error {
	statement := "INSERT INTO participants (name, participating) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET participating = $2"
	_, err := db.Exec(statement, participant.name, participant.participating)
	return err
}

func getParticipantsFromDatabase(db *sql.DB) ([]participant, error) {
	rows, err := db.Query("SELECT name, participating FROM participants")
	if err != nil {
		return nil, errors.Wrap(err, "failed to do query")
	}
	defer rows.Close()

	var participants []participant
	for rows.Next() {
		var p participant
		err = rows.Scan(&p.name, &p.participating)
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
