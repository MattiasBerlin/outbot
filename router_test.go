package main

import (
	"github.com/MattiasBerlin/outbot/commands"
	"github.com/MattiasBerlin/outbot/handlers"
	"testing"
)

func testRouter() Router {
	return Router{
		commands: make(map[string]*commands.Command),
		prefix:   "!",
		guildID:  "Dummy Guild ID",
		db:       nil,
	}
}

func TestAddAndGetCommand(t *testing.T) {
	testData := []struct {
		msg           string
		expectedTrail string
		cmd           commands.Command
	}{
		{msg: "event add", expectedTrail: "", cmd: handlers.EventAddCommand()},
		{msg: "event add 7m 321 123", expectedTrail: "7m 321 123", cmd: handlers.EventAddCommand()},
		{msg: "in", expectedTrail: "", cmd: handlers.EventAddCommand()},
		{msg: "in 3s something 123", expectedTrail: "3s something 123", cmd: handlers.EventAddCommand()},
		{msg: "event", expectedTrail: "", cmd: handlers.EventCommand()},
		{msg: "ping", expectedTrail: "", cmd: handlers.PingCommand()},
		{msg: "help event", expectedTrail: "event", cmd: handlers.HelpCommand()},
	}

	r := testRouter()
	cmds := getCommands()
	r.AddCommands(cmds)

	retreivedCmd, _ := r.getCommand("something which does not exist")
	if retreivedCmd != nil {
		t.Error("Should not return a command for a non-existing route")
	}

	for _, d := range testData {
		retrievedCmd, trail := r.getCommand(d.msg)
		if retrievedCmd == nil {
			t.Errorf("Should not return nil for %q", d.msg)
		} else {
			if retrievedCmd.CallPhrase != d.cmd.CallPhrase {
				t.Errorf("%q should return %q, not %q", d.msg, d.cmd.CallPhrase, retrievedCmd.CallPhrase)
			}
			if trail != d.expectedTrail {
				t.Errorf("Trail %q should be %q for message %q", trail, d.expectedTrail, d.msg)
			}
		}
	}
}
