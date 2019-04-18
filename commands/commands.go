package commands

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
)

type Permission int

const (
	All Permission = iota
	Members
	Officers
)

const (
	MemberRoleID  = "416353375647432706"
	AcademyRoleID = "488400983836196874"
	OfficerRoleID = "382256632882659338"
)

// Authorized returns whether the member is authorized to use the command.
func (p Permission) Authorized(user discordgo.Member) bool {
	authorized := false

	switch p {
	case All:
		authorized = true
	case Members:
		for _, r := range user.Roles {
			if r == MemberRoleID || r == AcademyRoleID || r == OfficerRoleID {
				authorized = true
				break
			}
		}
	case Officers:
		for _, r := range user.Roles {
			if r == OfficerRoleID {
				authorized = true
				break
			}
		}
	}

	return authorized
}

// Handler of message sent events. TODO: Jesus past me, this can't be the best way to do it
type Handler func(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []Command)
type Init func(s *discordgo.Session, db *sql.DB)

type Command struct {
	CallPhrase string
	// alternative callphrases TODO: always top-level?
	Aliases    []string
	Permission Permission
	// TODO: doc
	SubCommands     []Command
	HelpDescription string
	Handler         Handler
	// Init is called before the handler. Put it as nil if there's no need.
	Init Init
	Help Help
}

// Help with information about what the Command does and how to use it.
type Help struct {
	// Summary of what the command does in a short sentence.
	Summary string
	// DetailedDescription of what the command does.
	DetailedDescription string
	// Syntax shows how to use the command.
	// Leave this empty if there is no particular functionality in the main command, but instead in the subcommands.
	// Do not include a prefix (e.g. an exclamation mark).
	Syntax string
	// Example of how to use the command.
	// Leave this empty if there is no particular functionality in the main command, but instead in the subcommands.
	// Do not include a prefix (e.g. an exclamation mark).
	Example string
	// SubCommands contains the callphrase mapped to their own Help object.
	SubCommands map[string]Help // TODO: Finish implementing this, consider using a slice instead
}
