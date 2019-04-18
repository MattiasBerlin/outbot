package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/MattiasBerlin/outbot/database"
	"github.com/bwmarrin/discordgo"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	prefix  = "!"
	guildID = "382256124604448768" // TODO: Move to config
)

var (
	configPath string
)

func readFlags() {
	configPath = *flag.String("config", "", "Path to the directory containing the config file.")
}

func main() {
	apiKey := os.Getenv("OB_APIKEY")
	if apiKey == "" {
		panic("OB_APIKEY has to be set")
	}

	readFlags()
	// readConfig(configDir(configPath))

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

// configDir returns the config directory.
func configDir(flagDir string) string {
	if flagDir != "" {
		return flagDir
	}

	// XDG Base Directory Specification
	// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return xdgConfig
	}

	home := os.Getenv("HOME")
	return filepath.Join(home, ".config")
}

func logOutput(dir string) (io.Writer, error) {
	base := filepath.Join(filepath.Dir(dir), "outbot-log")
	file, err := rotatelogs.New(base+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(base),
		rotatelogs.WithMaxAge(time.Hour*24*30),    // keep 30 days
		rotatelogs.WithRotationTime(time.Hour*24), // rotate once a day
	)
	if err != nil {
		return nil, err
	}

	return io.MultiWriter(os.Stdout, file), nil
}
