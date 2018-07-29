package router

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/MattiasBerlin/outbot/handlers"
	"github.com/bwmarrin/discordgo"
	"strings"
)

// Router for commands.
type Router struct {
	commands map[string]*commands.Command
	prefix   string
	guildID  string
	db       *sql.DB
}

// New router. Adds and initializes the commands.
func New(prefix string, guildID string, s *discordgo.Session, db *sql.DB) *Router {
	r := &Router{
		commands: make(map[string]*commands.Command),
		prefix:   prefix,
		guildID:  guildID,
		db:       db,
	}

	cmds := r.getCommands()

	// Init commands
	for _, cmd := range cmds {
		if cmd.Init != nil {
			fmt.Println("Initializing handler:", cmd.CallPhrase)
			cmd.Init(s, db)
		}
	}

	r.AddCommands(cmds)
	return r
}

// AddCommand to the router.
func (r *Router) AddCommand(cmd commands.Command) {
	r.commands[cmd.CallPhrase] = &cmd
}

// AddCommands to the router.
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

// OnMessageSent gets called when a message is sent.
func (r *Router) OnMessageSent(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, r.prefix) {
		return
	}

	// Only get first word
	msg := strings.Split(m.Content, " ")[0]
	// Strip prefix
	msg = msg[len(r.prefix):]
	command := r.getCommand(msg)

	if command == nil {
		return
	}
	if command.Handler == nil {
		fmt.Println(command.CallPhrase, "does not have a handler")
		return
	}

	user, err := s.GuildMember(r.guildID, m.Author.ID)
	if err != nil {
		fmt.Println("Failed to obtain guild member:", err.Error())
		return
	}
	if user == nil {
		// Not sent by a user?
		// Probably a join message or something like that
		return
	}

	if command.Permission.Authorized(*user) {
		command.Handler(s, m, r.db, r.getAllCommands())
	} else {
		fmt.Println(m.Author.Username, "tried to use", command.CallPhrase, "without the required authorization")
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
		handlers.OptInCommand(),
		handlers.OptOutCommand(),
		handlers.ListParticipantsCommand(),
		handlers.ClearParticipantsCommand(),
	}
}
