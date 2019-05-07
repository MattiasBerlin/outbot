package handlers

// import (
// 	"database/sql"
// 	"fmt"
// 	"github.com/MattiasBerlin/outbot/commands"
// 	"github.com/bwmarrin/discordgo"
// 	"golang.org/x/oauth2/google"
// 	"io/ioutil"
// )

const (
	googleAPICredentialsFile = "credentials.json"
	spreadsheetID            = ""
)

// func InitMod(s *discordgo.Session, db *sql.DB) {
// 	b, err := ioutil.ReadFile(googleAPICredentialsFile)
// 	if err != nil {
// 		fmt.Println("Unable to read Google API credentials file:", err.Error())
// 		return
// 	}

// 	// If modifying these scopes, delete your previously saved token.json
// 	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
// 	if err != nil {
// 		fmt.Println("Unable to parse Google API credentials file to config:", err.Error())
// 	}
// }

// // EventCommand for reminders.
// func ModCommand() commands.Command {
// 	return commands.Command{
// 		CallPhrase:      "mods",
// 		Permission:      commands.Members,
// 		HelpDescription: "Get member's module levels",
// 		Handler:         HandleEvent,
// 		Init:            InitEvent,
// 		Help: commands.Help{
// 			Summary:             "Get member's module levels",
// 			DetailedDescription: "<TODO>",
// 		},
// 	}
// }
