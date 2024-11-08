package slackapp_test

import (
	"context"
	"github.com/clambin/slackapp"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"log/slog"
	"os"
	"os/signal"
)

func ExampleSlackApp() {
	const (
		slackToken = "xoxb-token"
		appToken   = "xapp-token"
	)
	c := slack.New(slackToken, slack.OptionAppLevelToken(appToken))
	app := slackapp.NewSlackApp(c, slog.Default())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	go func() {
		_ = app.Run(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-app.Events:
			switch ev.Data.(type) {
			case *slackevents.AppMentionEvent:
				// process app mention event
				// app.Client gives you access to the underlying slack client to post, or any other Slack functionality you may require.
				// You may need to add additional OAuth scopes to grant access to those APIs.
			case *slackevents.MessageEvent:
				// process direct message event
			}
		}
	}
}
