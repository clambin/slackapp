package slackapp

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"io"
	"log/slog"
	"regexp"
	"strings"
)

// Bot is a SlackApp application that receives commands by mentioning the bot in a channel. The bot executes the commands
// and posts the output in the channel where it was mentioned.
type Bot struct {
	*SlackApp
	Commands
	logger *slog.Logger
}

// NewBot creates a Bot for the Slack client.
func NewBot(client *slack.Client, options ...BotOptionFunc) *Bot {
	b := makeBot(options...)
	b.SlackApp = NewSlackApp(client, b.logger.With("component", "slackapp"))
	return b
}

func newBotWith(c *slack.Client, h socketModeHandler, options ...BotOptionFunc) *Bot {
	b := makeBot(options...)
	b.SlackApp = newSlackAppWithSocketModeHandler(socketmode.New(c), h, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return b
}

func makeBot(options ...BotOptionFunc) *Bot {
	b := Bot{
		Commands: make(Commands),
		logger:   slog.Default(),
	}
	for _, o := range options {
		o(&b)
	}
	return &b
}

// Run starts the bot. It connects to Slack and waits for a command. It executes the command and posts the output in the channel
// where the command was issued.
func (b *Bot) Run(ctx context.Context) error {
	botUserID, err := b.userID()
	if err != nil {
		return err
	}

	b.logger.Debug("starting Bot")
	defer b.logger.Debug("shutting down Bot")
	errCh := make(chan error)
	go func() { errCh <- b.SlackApp.Run(ctx) }()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err = <-errCh:
			if err != nil {
				err = fmt.Errorf("slackapp failed: %w", err)
			}
			return err
		case ev := <-b.SlackApp.Events:
			switch data := ev.Data.(type) {
			case *slackevents.AppMentionEvent:
				_ = b.handle(ctx, data.Channel, data.Text)
			case *slackevents.MessageEvent:
				// don't process our own messages
				if data.User != botUserID {
					_ = b.handle(ctx, data.Channel, data.Text)
				}
			default:
				b.logger.Warn("received unexpected Event API event", "type", ev.Type)
			}
		}
	}
}

func (b *Bot) handle(ctx context.Context, channel string, input string) error {
	args := tokenizeText(removeUserID(input))
	b.logger.Debug("executing command", "channel", channel, "cmd", args[0])
	resp := b.Handle(ctx, args...)
	_, _, err := b.SlackApp.Client.PostMessage(channel, resp...)
	return err
}

func (b *Bot) userID() (string, error) {
	auth, err := b.SlackApp.AuthTest()
	if err != nil {
		return "", fmt.Errorf("auth: %w", err)
	}
	return auth.UserID, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// BotOptionFunc configures a Bot when it's created.
type BotOptionFunc func(*Bot)

// WithLogger adds the provided logger to the Bot. The default is slog.Default().
func WithLogger(logger *slog.Logger) BotOptionFunc {
	return func(bot *Bot) {
		bot.logger = logger
	}
}

// WithCommand registers a command with the bot.
func WithCommand(verb string, handler Handler) BotOptionFunc {
	return func(bot *Bot) {
		bot.Commands[verb] = handler
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////f////

var tokenizerRegExp = regexp.MustCompile(`[^\s"]+|"([^"]*)"`)

func tokenizeText(input string) []string {
	cleanInput := input
	for _, quote := range []string{"“", "”", "'"} {
		cleanInput = strings.ReplaceAll(cleanInput, quote, "\"")
	}
	output := tokenizerRegExp.FindAllString(cleanInput, -1)

	for index, word := range output {
		output[index] = strings.Trim(word, "\"")
	}
	return output
}

var userIDRegExp = regexp.MustCompile(`<@\w+> (.*)$`)

func removeUserID(input string) string {
	matches := userIDRegExp.FindStringSubmatch(input)
	if len(matches) != 2 {
		return input
	}
	return matches[1]
}
