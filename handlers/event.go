package handlers
//vong testing
import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	botEventChannelID = "466576270285602823"

	eventExpiredColor = 0x4286f4
)

type event struct {
	description string
	time        time.Time
	expired     bool
}

// EventCommand for reminders.
func EventCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "event",
		Permission:      commands.Members,
		HelpDescription: "Set reminders, useful for WS",
		SubCommands: []commands.Command{
			EventAddCommand(),
		},
		Handler: HandleEvent,
		Init:    InitEvent,
		Help: commands.Help{
			Summary: "Set reminders, useful for WS",
			DetailedDescription: "Set reminders for events that will occur after a specific duration.\n" +
				"Available subcommands are: `add`, `upcoming` and `history` (*I'll add functionality to get help on subcommands too soon*)",
		},
	}
}

func EventAddCommand() commands.Command {
	return commands.Command{
		CallPhrase:      "add",
		Aliases:         []string{"in"},
		Permission:      commands.Members,
		HelpDescription: "Add a reminder",
		Handler:         HandleAddEvent,
		Help: commands.Help{
			Summary:             "Add a reminder",
			DetailedDescription: "Add a reminder.",
			Syntax:              "!event add <duration> <message>",
			Example:             "!event add 1h5m Write a message here",
		},
	}
}

func (e event) timeDBFormat() string {
	return e.time.Format("2006-01-02 15:04:05")
}

func InitEvent(s *discordgo.Session, db *sql.DB) {
	events, err := getEventsFromDatabase(db, 0, false)
	if err != nil {
		fmt.Println("Failed to get events from database on init:", err.Error())
		return
	}

	// Check if any events should have went off while the bot was offline, otherwise set a timer
	var missedEvents string
	for _, e := range events {
		if e.time.Before(time.Now()) {
			// Event should have went off while bot was offline
			err = setEventExpiredInDatabase(db, e, true)
			if err != nil {
				fmt.Println("Failed to set event expired in db:", err.Error())
			}
			missedEvents += fmt.Sprintf("* %v ago: %q\n", e.time.String()[1:], e.description)
		} else {
			startEventTimer(e, s, db)
			fmt.Println(fmt.Sprintf("Started timer for %q", e.description))
		}
	}
	if missedEvents != "" {
		_, err := s.ChannelMessageSend(botEventChannelID, fmt.Sprintf("Events expired while bot was offline:\n%v", missedEvents))
		if err != nil {
			fmt.Println("Failed to send message:", err)
		}
	}
}

func HandleEvent(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	split := strings.Split(msg, " ")

	switch split[1] {
	case "add":
		if len(split) < 2 {
			msg := discordgo.MessageEmbed{
				Title:       "Incorrect syntax",
				Color:       failColor,
				Description: "Too few arguments for add event command, check `!help event`",
			}
			_, err := s.ChannelMessageSendEmbed(m.ChannelID, &msg)
			if err != nil {
				fmt.Println("Failed to send message:", err)
			}
			return
		}

		err := addEvent(s, m, split, db)
		if err != nil {
			fmt.Println("Failed to add event:", err.Error())
			return
		}
	case "upcoming":
		upcoming, err := getEventsFromDatabase(db, 10, false)
		if err != nil {
			fmt.Println("Failed to get upcoming events:", err.Error())
			_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to get events: %v", err))
			if err != nil {
				fmt.Println("Failed to send message:", err.Error())
				return
			}
			return
		}

		// content := "**Upcoming events:**\n"
		var content string
		for _, e := range upcoming {
			content += fmt.Sprintf("* In %v: %v\n", e.time.Sub(time.Now()).Round(time.Second), e.description)
		}

		msg := discordgo.MessageEmbed{
			Title:       "Upcoming events",
			Color:       infoColor,
			Description: content,
		}
		_, err = s.ChannelMessageSendEmbed(m.ChannelID, &msg)
		if err != nil {
			fmt.Println("Failed to send message:", err.Error())
			return
		}
	case "history":
		pastEvents, err := getEventsFromDatabase(db, 10, true)
		if err != nil {
			fmt.Println("Failed to get past events:", err.Error())
			_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to get events: %v", err))
			if err != nil {
				fmt.Println("Failed to send message:", err.Error())
				return
			}
			return
		}

		var content string
		for _, e := range pastEvents {
			content += fmt.Sprintf("* %v ago: %v\n", e.time.Sub(time.Now()).Round(time.Second).String()[1:], e.description) // TODO: Pretty this
		}

		msg := discordgo.MessageEmbed{
			Title:       "Past events",
			Color:       infoColor,
			Description: content,
		}
		_, err = s.ChannelMessageSendEmbed(m.ChannelID, &msg)
		if err != nil {
			fmt.Println("Failed to send message:", err.Error())
			return
		}
	default:
		msg := discordgo.MessageEmbed{
			Title:       "Incorrect syntax",
			Color:       failColor,
			Description: "Unknown event command, check `!help event`",
		}
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, &msg)
		if err != nil {
			fmt.Println("Failed to send message:", err.Error())
			return
		}
		// delete
	}
}

func HandleAddEvent(msg string, s *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, guildID string, cmds []commands.Command) {
	split := strings.Split(msg, " ") // TODO: Won't work, pass trailing message as parameter

	if len(split) < 2 {
		msg := discordgo.MessageEmbed{
			Title:       "Incorrect syntax",
			Color:       failColor,
			Description: "Too few arguments for add event command, check `!help event`",
		}
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, &msg)
		if err != nil {
			fmt.Println("Failed to send message:", err)
		}
		return
	}

	err := addEvent(s, m, split, db)
	if err != nil {
		fmt.Println("Failed to add event:", err.Error())
		return
	}
}

func addEvent(s *discordgo.Session, m *discordgo.MessageCreate, splitMsg []string, db *sql.DB) error {
	duration, err := time.ParseDuration(splitMsg[0])
	if err != nil {
		msg := discordgo.MessageEmbed{
			Title:       "Incorrect syntax",
			Color:       failColor,
			Description: "Incorrect syntax for duration, check `!help event`",
		}
		_, err = s.ChannelMessageSendEmbed(m.ChannelID, &msg)
		return errors.Wrap(err, "failed to send message")
	}

	event := event{
		description: strings.Join(splitMsg[1:], " "),
		time:        time.Now().Add(duration),
	}

	err = addEventToDatabase(db, event)
	if err != nil {
		fmt.Println("Failed to add event:", err.Error())
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to add event: %v", err))
		return errors.Wrap(err, "failed to send message")
	}

	startEventTimer(event, s, db)

	msg := discordgo.MessageEmbed{
		Title:       "Event added!",
		Color:       successColor,
		Description: fmt.Sprintf("In %v: %q", duration.String(), event.description),
	}
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, &msg)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}

func startEventTimer(event event, s *discordgo.Session, db *sql.DB) {
	duration := event.time.Sub(time.Now())
	timer := time.NewTimer(duration)
	go waitForEventTimerExpire(event, timer.C, s, db)
}

func waitForEventTimerExpire(event event, c <-chan time.Time, s *discordgo.Session, db *sql.DB) {
	<-c
	fmt.Println(event.description, "expired")

	err := setEventExpiredInDatabase(db, event, true)
	if err != nil {
		fmt.Println("Failed to set event expired in db:", err.Error())
	}

	msg := discordgo.MessageEmbed{
		Title:       "Event expired",
		Color:       infoColor,
		Description: event.description,
	}
	_, err = s.ChannelMessageSendEmbed(botEventChannelID, &msg)
	if err != nil {
		fmt.Println("Failed to send message:", err.Error())
		return
	}
}

func addEventToDatabase(db *sql.DB, event event) error {
	_, err := db.Exec("INSERT INTO events (description, time) VALUES ($1, $2)", event.description, event.timeDBFormat())
	return err
}

// getEventsFromDatabase.
// If limit is <=0 then no limit will be used.
func getEventsFromDatabase(db *sql.DB, limit int, expired bool) ([]event, error) {
	query := "SELECT description, time, expired FROM events WHERE expired = $1 ORDER BY time ASC"
	args := []interface{}{expired}
	if limit > 0 {
		query += " LIMIT $2"
		args = append(args, limit)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do query")
	}
	defer rows.Close()

	upcoming := make([]event, 0, limit)

	for rows.Next() {
		var event event
		err = rows.Scan(&event.description, &event.time, &event.expired)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		upcoming = append(upcoming, event)
	}

	return upcoming, nil
}

func setEventExpiredInDatabase(db *sql.DB, e event, expired bool) error {
	_, err := db.Exec("UPDATE events SET expired = $1 WHERE description = $2 AND time = $3", expired, e.description, e.timeDBFormat())
	return errors.Wrap(err, "failed to execute query")
}
