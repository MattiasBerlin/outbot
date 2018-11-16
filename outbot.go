package main

import (
	"database/sql"
	"fmt"
	"github.com/MattiasBerlin/outbot/database"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
)

const (
	prefix  = "!"
	guildID = "382256124604448768" // TODO: Remove
)

func main() {
	apiKey := os.Getenv("OB_APIKEY")
	if apiKey == "" {
		panic("OB_APIKEY has to be set")
	}

	session, err := discordgo.New("Bot " + apiKey)
	if err != nil {
		fmt.Println("Failed to create discord session:", err)
		return
	}

	db, err := connectToDatabase()
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
	}

	router := NewRouter(prefix, guildID, session, db)

	session.AddHandler(router.OnMessageSent)

	err = session.Open()
	if err != nil {
		fmt.Println("Failed to open session:", err)
		return
	}

	fmt.Println("Up and running! Press CTRL+C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	session.Close()
}

func connectToDatabase() (*sql.DB, error) {
	db, err := database.New()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open connection")
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping db")
	}

	return db, nil
}

// func onMessageSent(s *discordgo.Session, m *discordgo.MessageCreate) {
// 	if m.Author.Bot {
// 		return
// 	}

// 	if m.Content == "!test" {
// 		_, err := s.ChannelMessageSend(m.ChannelID, "It's working!")
// 		if err != nil {
// 			fmt.Println("Failed to send message: ", err)
// 			return
// 		}
// 	}
// }
