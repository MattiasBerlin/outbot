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
			if r == MemberRoleID || r == OfficerRoleID {
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

type Handler func(s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []Command)
type Init func(s *discordgo.Session, db *sql.DB)

type Command struct {
	CallPhrase      string
	Permission      Permission
	HelpDescription string
	Handler         Handler
	// Init is called before the handler. Put is as nil if there's no need.
	Init Init
}
