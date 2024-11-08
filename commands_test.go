package slackapp

import (
	"context"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCommands(t *testing.T) {
	handler := func(text string) Handler {
		return HandlerFunc(func(_ context.Context, args ...string) []slack.MsgOption {
			if len(args) > 0 {
				text += ": " + strings.Join(args, ", ")
			}
			return []slack.MsgOption{slack.MsgOptionText(text, true)}
		})
	}

	tests := []struct {
		name     string
		commands Commands
		args     []string
		want     map[string]string
	}{
		{
			name:     "single command",
			commands: Commands{"foo": handler("foo")},
			args:     []string{"foo"},
			want:     map[string]string{"text": "foo"},
		},
		{
			name:     "single command with args",
			commands: Commands{"foo": handler("foo")},
			args:     []string{"foo", "a=b"},
			want:     map[string]string{"text": "foo: a=b"},
		},
		{
			name:     "empty",
			commands: Commands{"foo": handler("foo")},
			args:     nil,
			want:     map[string]string{"attachments": `[{"color":"bad","title":"invalid command","text":"supported commands: foo","blocks":null}]`},
		},
		{
			name:     "invalid command",
			commands: Commands{"foo": handler("foo")},
			args:     []string{"bar"},
			want:     map[string]string{"attachments": `[{"color":"bad","title":"invalid command","text":"supported commands: foo","blocks":null}]`},
		},
		{
			name:     "nested command",
			commands: Commands{"foo": &Commands{"bar": handler("bar")}},
			args:     []string{"foo", "bar"},
			want:     map[string]string{"text": "bar"},
		},
		{
			name:     "invalid nested command",
			commands: Commands{"foo": &Commands{"bar": handler("bar")}},
			args:     []string{"foo", "foo"},
			want:     map[string]string{"attachments": `[{"color":"bad","title":"invalid command","text":"supported commands: bar","blocks":null}]`},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := make(Commands)
			c.Add(tt.commands)
			options := c.Handle(context.Background(), tt.args...)
			// slack.MsgOption doesn't have a way to return the formatted message.
			// so we need to get it in a convoluted way
			output := formatMessage(options)

			for k, v := range tt.want {
				require.Contains(t, output, k)
				assert.Equal(t, v, output.Get(k))
			}
		})
	}
}

// formatMessage formats a set of slack.MsgOption items.  Since the sdk doesn't give a direct way of formatting them,
// we do it in a convoluted way: we start a fake server, post a message with the slack.MsgOption items and capture the request.
func formatMessage(options []slack.MsgOption) url.Values {
	var values url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body of the request (the JSON payload)
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		values, _ = url.ParseQuery(string(body))
		// Respond with a dummy success response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	api := slack.New("xoxb-your-slack-bot-token", slack.OptionAPIURL(server.URL+"/"))
	// Post message to a channel (will hit our mock server instead of Slack)
	_, _, err := api.PostMessage("ChannelID", options...)
	if err != nil {
		log.Fatalf("Error posting message: %v", err)
	}

	return values
}
