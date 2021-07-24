package main

import (
	"log"
	"os"

	cron "github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	yaml "gopkg.in/yaml.v3"
)

const slackClient = slack.New(getEnvWithError("ROTABOT_SLACK_TOKEN"), slack.OptionDebug(getEnvWithDefault("ROTABOT_DEBUG", "0")))
const userGroup = getEnvWithError("ROTABOT_USERGROUP")

type Rota struct {
	Users []string `yaml:users`
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func getEnvWithDefault(env string, def string) string {
	envValue, ok := os.LookupEnv(env)

	if !ok {
		envValue = def
	}

	return envValue
}

func getEnvWithError(env string) string {
	envValue, ok := os.LookupEnv(env)

	if !ok {
		log.Fatalf("Missing required environment variable: %s", env)
	}

	return envValue
}

func seekNextUser(email string, users []string) string {
	totalUsers := len(users)

	for i, userEmail := range users {
		if userEmail == email {
			// get the next user in the list - wrap around if the current user is that last user in the list
			return users[(i+1)%totalUsers]
		}
	}

	// if the current user is not found, will start from top of the list
	return users[0]
}

func rotate() {
	var rota Rota

	//go:embed rota.yaml
	var inRota []byte

	err := yaml.Unmarshal(inRota, &rota)
	check(err, "Could not unmarshal rota.yaml")

	groupMembers, err := slackClient.GetUserGroupMembers(userGroup)

	if err != nil {
		// List of Slack error responses: https://api.slack.com/methods/usergroups.users.list
		switch {
		case err.Error() == "no_such_subteam":
			// TODO: finish create usergroup functionality (need function to create the usergroup)
			slackClient.CreateUserGroup(userGroup)
		default:
			check(err, "Could not get group members")
		}
	}

	var todaySupportUserEmail string

	// if @prodeng-support not set or has more than 1 member, rotaboton will set the group to the first user in rota.yaml; otherwise
	// it will start the rotation from the currently set user (if that currently set user is not in the rota list, the rotiation will
	// start from the first user in rota.yaml)
	if len(groupMembers) != 1 {
		todaySupportUserEmail = rota.Users[0]
	} else {
		user, err := slackClient.GetUserInfo(groupMembers)

		check(err, "Could not get user information")

		userEmail := user.Profile.Email

		todaySupportUserEmail = seekNextUser(userEmail, rota.Users)
	}

	todaySupportUserName, err := slackClient.GetUserByEmail(todaySupportUserEmail)
	check(err, "Could not get user by email")

	slackClient.UpdateUserGroupMembers(userGroup, todaySupportUserName)
}

func main() {
	scheduler := cron.New()

	scheduler.AddFunc("0 0 0 * * 1-5", rotate)
}
