package slackapp_test

import (
	"context"
	"github.com/clambin/slackapp"
	"github.com/slack-go/slack"
)

func ExampleBot() {
	const (
		slackToken = "xoxb-token"
		appToken   = "xapp-token"
	)
	c := slack.New(slackToken, slack.OptionAppLevelToken(appToken))

	b := slackapp.NewBot(c,
		// this registers a command "foo", that posts "bar" as a response. See slack.MsgOption for possible outputs.
		slackapp.WithCommand("foo", slackapp.HandlerFunc(func(ctx context.Context, args ...string) []slack.MsgOption {
			return []slack.MsgOption{slack.MsgOptionText("bar", false)}
		})),
	)

	_ = b.Run(context.Background())
}
