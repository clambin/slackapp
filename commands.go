package slackapp

import (
	"context"
	"github.com/slack-go/slack"
	"slices"
	"strings"
)

// A Handler executes a command and returns messages to be posted to Slack.
type Handler interface {
	Handle(context.Context, ...string) []slack.MsgOption
}

// HandlerFunc is an adapter that allows a function to be used as a Handler
type HandlerFunc func(context.Context, ...string) []slack.MsgOption

// Handle calls f(ctx, args)
func (f HandlerFunc) Handle(ctx context.Context, args ...string) []slack.MsgOption {
	return f(ctx, args...)
}

var _ Handler = Commands{}

// Commands is a map of verb/Handler pairs.
//
// Note that Commands itself implements the Handler interface. This allows nested command structures to be built:
//
//	Commands
//	"foo"    -> handler
//	"bar"    -> Commands
//	            "snafu"    -> handler
//
// This creates the commands "foo" and "bar snafu"
type Commands map[string]Handler

func (c Commands) Handle(ctx context.Context, args ...string) []slack.MsgOption {
	if cmd, params := split(args...); cmd != "" {
		if subCommand, ok := c[cmd]; ok {
			return subCommand.Handle(ctx, params...)
		}
	}

	return []slack.MsgOption{slack.MsgOptionAttachments(slack.Attachment{
		Color: "bad",
		Title: "invalid command",
		Text:  "supported commands: " + strings.Join(c.GetCommands(), ", "),
	})}
}

func split(args ...string) (string, []string) {
	if len(args) == 0 {
		return "", nil
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return args[0], args[1:]
}

// GetCommands returns a sorted list of all supported commands.
func (c Commands) GetCommands() []string {
	commands := make([]string, 0, len(c))
	for verb := range c {
		commands = append(commands, verb)
	}
	slices.Sort(commands)
	return commands
}

// Add adds one or more commands.
func (c Commands) Add(commands Commands) {
	for verb, handler := range commands {
		c[verb] = handler
	}
}
