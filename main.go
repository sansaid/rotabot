package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/slack-go/slack"
)

var slackDebug bool

func init() {
	if getEnvWithDefault("ROTABOT_DEBUG", "0") == "0" {
		slackDebug = false
	} else {
		slackDebug = true
	}

}

var slackToken string = getEnvWithError("ROTABOT_SLACK_TOKEN")
var slackClient *slack.Client = slack.New(slackToken, slack.OptionDebug(slackDebug), slack.OptionLog(&log.Logger{}))
var rotaBot *RotaBot = NewRotaBot()

func handleEmptySubcommand(w http.ResponseWriter, _ *Command) {
	msg := NewSlackMsg()

	msg.Text = fmt.Sprintf("Valid subcommands: %s", rotaBot.ListValidCommands())

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func handleDefault(w http.ResponseWriter, p *Command) {
	msg := NewSlackMsg()

	msg.Text = fmt.Sprintf("Invalid subcommand: %s. Valid subcommands: %s", p.Subcommand, rotaBot.ListValidCommands())

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func slashHandler(w http.ResponseWriter, r *http.Request) {
	s, err := slack.SlashCommandParse(r)
	check(err, "Could not parse Slash command")

	p, err := parse(s.Command, s.Text)
	check(err, "Could not parse Slash command")

	switch p.Subcommand {
	case "":
		handleEmptySubcommand(w, p)
	default:
		handleDefault(w, p)
	}
}

func main() {
	http.HandleFunc("/", slashHandler)

	http.ListenAndServe(":8080", nil)
}
