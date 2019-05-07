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

	academyWhiteStarChannel   = "488859067947941909"
	academyGeneralChannel     = "488401533063659530"
	academyMessagesChannel    = "488478592297336834"
	academyOfficersChannel    = "512368438266691594"
	academySpreadsheetChannel = "489510100705214464"
)

type wsRole string

const (
	defaultRole wsRole = "No preference"
	defense     wsRole = "Defense"
	offense     wsRole = "Offense"
	hunter      wsRole = "Hunter"
	filler      wsRole = "Filler"
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
	case "filler", "fill":
		role = filler
	default:
		role = defaultRole
	}
	return role
}

type instance string

const (
	mainA    instance = "Main A"
	mainB    instance = "Main B"
	academyA instance = "Academy A"
	academyB instance = "Academy B"
)

func secondInstance(instanceString string) bool {
	instanceString = strings.ToLower(instanceString)
	return instanceString == "b" || instanceString == "2"
}

func channelToInstance(channelID string, instanceString string) instance {
	var instance instance

	switch channelID {
	case academyWhiteStarChannel,
		academyGeneralChannel,
		academyMessagesChannel,
		academyOfficersChannel,
		academySpreadsheetChannel:
		if secondInstance(instanceString) {
			instance = academyB
		} else {
			instance = academyA
		}
	default:
		if secondInstance(instanceString) {
			instance = mainB
		} else {
			instance = mainA
		}
	}

	return instance
}

type participant struct {
	instance
	name          string
	participating bool
	preferredRole wsRole
	userID        string
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
			DetailedDescription: `Opt in for the next White Star match.
				Instances: A or B
				Accepted preferred roles: def, off, hunter`,
			Syntax:  "optin [instance] [preferred role]",
			Example: "optin A defense",
		},
	}
}

// SetOptInCommand for opting in other members to white stars.
func SetOptInCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "setoptin",
		Permission:      commands.Officers,
		HelpDescription: "Opt in other members for the next WS",
		Handler:         HandleSetOptIn,
		Help: commands.Help{
			Summary:             "Opt in other members for the next WS",
			DetailedDescription: `Opt in other members for the next White Star match.`,
			Syntax:              "setoptin [instance] [role] <members>",
			Example:             "setoptin A def @Maro",
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
			Syntax:              "optout [instance]",
			Example:             "optout B",
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
			Syntax:              "clear [instance]",
			Example:             "clear A",
		},
	}
}

// HandleSetOptIn handles opt in commands for mentioned users.
func HandleSetOptIn(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	splitMsg := strings.Split(msg, " ")
	var instanceString string
	role := defaultRole
	if len(splitMsg) >= 1 {
		instanceString = splitMsg[0]
		if len(splitMsg) >= 2 {
			role = wsRoleFromString(splitMsg[1])
		}
	}
	instance := channelToInstance(m.ChannelID, instanceString)

	message := fmt.Sprintf("You've opted in %d members.", len(m.Mentions))

	for i, user := range m.Mentions {
		if user != nil {
			m.Author = user
			setParticipation(true, instance, role, message, i == len(m.Mentions)-1, s, m, db)
		}
	}
}

// HandleOptIn handles opt in commands.
func HandleOptIn(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	splitMsg := strings.Split(msg, " ")
	instance := splitMsg[0]
	var wsRole string
	if len(splitMsg) >= 2 {
		wsRole = splitMsg[1]
	}
	setParticipation(true, channelToInstance(m.ChannelID, instance), wsRoleFromString(wsRole), fmt.Sprintf("You've opted in, %v!", m.Author.Username), true, s, m, db)
}

// HandleOptOut handles opt out commands.
func HandleOptOut(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	setParticipation(false, channelToInstance(m.ChannelID, msg), wsRoleFromString(""), fmt.Sprintf("You've opted out, %v!", m.Author.Username), true, s, m, db)
}

// HandleClearParticipants handles clearing the participation list.
func HandleClearParticipants(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	instance := channelToInstance(m.ChannelID, msg)
	rolesRemoved := removeRolesForParticipants(instance, s, db, guildID)
	err := clearParticipantsFromDatabase(db, instance)
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
		Description: fmt.Sprintf("Participation list cleared!\nCleared roles from %d members.", rolesRemoved),
	}
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &response)
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}

// HandleListParticipants handles the command for listing participants.
func HandleListParticipants(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	listParticipants("", channelToInstance(m.ChannelID, msg), s, m, db)
}

func listParticipants(prefix string, instance instance, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB) {
	status, err := optStatus(db, instance)
	if err != nil {
		fmt.Println("Failed to get participation status:", err.Error())
		status = "[Failed to get participation status]"
	}

	response := discordgo.MessageEmbed{
		Color:       successColor,
		Description: prefix + status,
	}
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &response)
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}

func setParticipation(participating bool, instance instance, preferredRole wsRole, updateMessage string, sendMessage bool, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB) {
	participant := participant{
		instance:      instance,
		name:          m.Author.Username,
		participating: participating,
		preferredRole: preferredRole,
		userID:        m.Author.ID,
	}
	err := setParticipatingInDatabase(db, participant)
	if err != nil {
		fmt.Println("Failed to set participation:", err.Error())
		return
	}
	if sendMessage {
		listParticipants(fmt.Sprintf("%v\n\n", updateMessage), participant.instance, s, m, db)
	}
}

func optStatus(db *sql.DB, instance instance) (string, error) {
	participants, err := getParticipantsFromDatabase(db, instance)
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

	var (
		roles       string
		fillerNames string
		fillerCount int
	)
	for role, names := range roleMap {
		// Filler is special, not counted in the total opted in
		if role == string(filler) {
			fillerCount = len(names)
			fillerNames = strings.Join(names, ", ")
		} else {
			roles += fmt.Sprintf("*%v* (%v): %v\n", role, len(names), strings.Join(names, ", "))
		}
	}
	roles += fmt.Sprintf("**Filler** (%v): %v\n", fillerCount, fillerNames)

	fillerCountText := ""
	if fillerCount > 0 {
		fillerCountText = fmt.Sprintf(" + %d filler", fillerCount)
		if fillerCount > 1 {
			fillerCountText += "s"
		}
	}

	return fmt.Sprintf("**Participants in %v**:\n**Opted in** (%d%s):\n%s**Opted out** (%d):\n%s", instance, len(optIn)-fillerCount, fillerCountText, roles, len(optOut), strings.Join(optOut, ", ")), nil
}

func setParticipatingInDatabase(db *sql.DB, participant participant) error {
	statement := `INSERT INTO participants (instance, name, participating, preferred_role, user_id) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (instance, name) DO UPDATE SET participating = $3, preferred_role = $4`
	_, err := db.Exec(statement, participant.instance, participant.name, participant.participating, participant.preferredRole, participant.userID)
	return err
}

func getParticipantsFromDatabase(db *sql.DB, instance instance) ([]participant, error) {
	rows, err := db.Query("SELECT name, participating, preferred_role, user_id FROM participants WHERE instance = $1", instance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do query")
	}
	defer rows.Close()

	var participants []participant
	for rows.Next() {
		var p participant
		err = rows.Scan(&p.name, &p.participating, &p.preferredRole, &p.userID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		participants = append(participants, p)
	}

	return participants, nil
}

func clearParticipantsFromDatabase(db *sql.DB, instance instance) error {
	_, err := db.Exec("DELETE FROM participants WHERE instance = $1", instance)
	return err
}
