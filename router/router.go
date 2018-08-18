package router

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/MattiasBerlin/outbot/handlers"
	"github.com/bwmarrin/discordgo"
	"sort"
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
	cmds := make([]commands.Command, len(r.commands))
	keys := make([]string, len(r.commands))

	// Go randomizes the element order if you use range on the map directly,
	// so the keys have to be sorted to get them in a consistent order
	i := 0
	for key := range r.commands {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	i = 0
	for _, key := range keys {
		cmds[i] = *r.commands[key]
		i++
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
		command.Handler(s, m, r.db, r.getAllCommands()) // TODO: There's no need to get all the commands every call, just do it once and save it
	} else {
		fmt.Println(m.Author.Username, "tried to use", command.CallPhrase, "without the required authorization")
	}
}

func (r *Router) getCommands() []commands.Command {
	return []commands.Command{
		handlers.HelpCommand(),
		handlers.PingCommand(),
		handlers.EventCommand(),
		handlers.OptInCommand(),
		handlers.OptOutCommand(),
		handlers.ListParticipantsCommand(),
		handlers.ClearParticipantsCommand(),
	}
}
