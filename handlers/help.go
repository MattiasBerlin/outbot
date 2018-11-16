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
	infoColor = 0x4286f4
)

// HelpCommand for getting help descriptions.
func HelpCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "help",
		Permission:      commands.All,
		HelpDescription: "Get descriptions of the available commands",
		Handler:         HandleHelp,
		Help: commands.Help{
			Summary:             "Get descriptions of the available commands",
			DetailedDescription: "Get descriptions of all the available commands.",
			Syntax:              "help [command]",
			Example:             "help ping",
		},
	}
}

// HandleHelp handles the help command.
func HandleHelp(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, cmds []commands.Command) {
	splitMsg := strings.Split(msg, " ")
	if len(splitMsg) == 0 {
		sendHelpMessage(s, m, cmds)
		return
	}

	for _, cmd := range cmds {
		if cmd.CallPhrase == splitMsg[0] {
			response := discordgo.MessageEmbed{
				Title: cmd.CallPhrase,
				Color: infoColor,
			}

			var content strings.Builder

			var desc string
			if cmd.Help.DetailedDescription != "" {
				desc = cmd.Help.DetailedDescription
			} else {
				desc = cmd.Help.Summary
			}
			content.WriteString(desc)

			hasSyntax := cmd.Help.Syntax != ""
			hasExample := cmd.Help.Example != ""
			if hasSyntax || hasExample {
				content.WriteString("\n\n")

				if hasSyntax {
					content.WriteString(fmt.Sprintf("Syntax: %v\n", cmd.Help.Syntax))
				}
				if hasExample {
					content.WriteString(fmt.Sprintf("Example: %v\n", cmd.Help.Example))
				}
			}

			response.Description = content.String()

			_, err := s.ChannelMessageSendEmbed(m.ChannelID, &response)
			if err != nil {
				fmt.Println("Failed to send message:", err)
				return
			}

			return
		}
	}

	// If it gets here no command was found matching the request
	response := discordgo.MessageEmbed{
		Color:       failColor,
		Description: fmt.Sprintf("Command %q was not found", msg[1]),
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &response)
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}
}

func sendHelpMessage(s *discordgo.Session, m *discordgo.MessageCreate, cmds []commands.Command) error {
	msg := discordgo.MessageEmbed{
		Title: "Command list",
		Color: infoColor,
	}

	var content string
	for _, cmd := range cmds {
		description := cmd.Help.Summary
		if description == "" {
			description = "*No description available*"
		}

		content += fmt.Sprintf("%v - %v\n", cmd.CallPhrase, description)
	}
	msg.Description = content

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &msg)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}
