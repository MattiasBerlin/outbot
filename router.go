package main

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

// NewRouter adds and initializes the commands.
func NewRouter(prefix string, guildID string, s *discordgo.Session, db *sql.DB) *Router {
	r := &Router{
		commands: make(map[string]*commands.Command),
		prefix:   prefix,
		guildID:  guildID,
		db:       db,
	}

	cmds := getCommands()

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
// Aliases are also registered.
func (r *Router) AddCommand(cmd commands.Command) {
	r.commands[cmd.CallPhrase] = &cmd

	// Add aliases
	for _, alias := range cmd.Aliases {
		// TODO: For now the prefix (super's callphrase or whatever) is ignored for aliases but
		// it would be nice to be able to use it if wanted
		r.commands[alias] = &cmd
	}

	r.addSubCommands(cmd)
}

func (r *Router) addSubCommands(cmd commands.Command) {
	for _, sub := range cmd.SubCommands {
		for _, alias := range sub.Aliases {
			r.commands[alias] = &sub
		}
		r.addSubCommands(sub)
	}
}

// AddCommands to the router.
func (r *Router) AddCommands(cmds []commands.Command) {
	for _, cmd := range cmds {
		r.AddCommand(cmd)
	}
}

// getCommand returns the command matching the message.
// The remaining text (after the command) is also returned.
func (r *Router) getCommand(msg string) (*commands.Command, string) {
	split := strings.Split(msg, " ")

	fmt.Println("Checking", split[0])
	cmd := r.commands[split[0]]

	// Check for subcommand matches if relevant
	if cmd != nil && len(split) > 1 {
		split = split[1:]
		sub := r.getSubCommand(cmd, split) // TODO: Ugly, make this prettier..
		for sub != nil {
			cmd = sub
			split = split[1:]
			sub = r.getSubCommand(sub, split)
		}
	} else {
		if len(split) > 1 {
			split = split[1:]
		} else {
			split = []string{}
		}
	}

	return cmd, strings.Join(split, " ")
}

func (r *Router) getSubCommand(cmd *commands.Command, trail []string) *commands.Command {
	for _, sub := range cmd.SubCommands {
		if sub.CallPhrase == trail[0] {
			return &sub
		}
	}

	return nil
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

// OnMessageSent gets called when a message is sent and routes to the correct handler based on the message.
func (r *Router) OnMessageSent(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, r.prefix) {
		return
	}

	// Only get first word
	//msg := strings.Split(m.Content, " ")[0]

	// Strip prefix
	msg := m.Content[len(r.prefix):]
	command, msg := r.getCommand(msg)
	if command == nil {
		fmt.Println("Command not found")
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
		command.Handler(msg, s, m, r.db, guildID, r.getAllCommands()) // TODO: There's no need to get all the commands every call, just do it once and save it
	} else {
		fmt.Println(m.Author.Username, "tried to use", command.CallPhrase, "without the required authorization")
	}
}

func getCommands() []commands.Command {
	return []commands.Command{
		handlers.HelpCommand(),
		handlers.PingCommand(),
		handlers.EventCommand(),
		handlers.OptInCommand(),
		handlers.OptOutCommand(),
		handlers.SetOptInCommand(),
		handlers.ListParticipantsCommand(),
		handlers.ClearParticipantsCommand(),
		handlers.StatusCommand(),
		handlers.SetRolesCommand(),
	}
}
