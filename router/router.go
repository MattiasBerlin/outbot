package router

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/MattiasBerlin/outbot/handlers"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type Router struct {
	commands map[string]*commands.Command
	prefix   string
	db       *sql.DB
}

func New(prefix string, s *discordgo.Session, db *sql.DB) *Router {
	r := &Router{
		commands: make(map[string]*commands.Command),
		prefix:   prefix,
		db:       db,
	}

	cmds := r.getCommands()

	// Init commands
	for _, cmd := range cmds {
		fmt.Println("asjidhasjdh", cmd.CallPhrase)
		if cmd.Init != nil {
			fmt.Println("Init", cmd.CallPhrase)
			cmd.Init(s, db)
		}
	}

	r.AddCommands(cmds)
	return r
}

func (r *Router) AddCommand(cmd commands.Command) {
	r.commands[cmd.CallPhrase] = &cmd
}

func (r *Router) AddCommands(cmds []commands.Command) {
	for _, cmd := range cmds {
		r.AddCommand(cmd)
	}
}

func (r *Router) getCommand(name string) *commands.Command {
	return r.commands[name]
}

func (r *Router) getAllCommands() []commands.Command {
	cmds := make([]commands.Command, 0, len(r.commands))

	for _, cmd := range r.commands {
		cmds = append(cmds, *cmd)
	}

	return cmds
}

func (r *Router) OnMessageSent(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, r.prefix) {
		return
	}

	// Only get first word
	msg := strings.Split(m.Content, " ")[0]
	// Strip prefix
	msg = msg[len(r.prefix):]
	//fmt.Println(msg)
	command := r.getCommand(msg)
	// TODO: Check permission!
	if command != nil {
		command.Handler(s, m, r.db, r.getAllCommands())
	}
}

func (r *Router) getCommands() []commands.Command {
	return []commands.Command{
		{
			CallPhrase:      "help",
			Permission:      commands.All,
			HelpDescription: "Get descriptions of the available commands",
			Handler:         handlers.HandleHelp,
		},
		{
			CallPhrase:      "ping",
			Permission:      commands.All,
			HelpDescription: "Check if OutBot is online",
			Handler:         handlers.HandlePing,
		},
		{
			CallPhrase:      "event",
			Permission:      commands.Members,
			HelpDescription: "Set reminders, useful for WS",
			Handler:         handlers.HandleEvent,
			Init:            handlers.InitEvent,
		},
	}
}
