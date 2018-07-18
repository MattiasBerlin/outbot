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
