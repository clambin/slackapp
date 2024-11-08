package slackapp

import (
	"context"
	"github.com/clambin/slackapp/internal/testutils"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestBot(t *testing.T) {
	ts := testServer{t: t, post: make(chan url.Values)}
	s := httptest.NewServer(&ts)
	defer s.Close()

	api := slack.New("x0xb-foo", slack.OptionAPIURL(s.URL+"/"))
	var h testutils.FakeHandler
	b := newBotWith(api, &h,
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithHandler("foo", HandlerFunc(func(ctx context.Context, s ...string) []slack.MsgOption {
			return []slack.MsgOption{slack.MsgOptionText("foo", false)}
		})),
	)

	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { errCh <- b.Run(ctx) }()

	// valid command
	slackClient := slack.New("", slack.OptionHTTPClient(&http.Client{Transport: &testutils.StubbedRoundTripper{}}))
	smClient := socketmode.New(slackClient)
	go b.SlackApp.SocketModeHandler.(*testutils.FakeHandler).SendEvent(testutils.AppMentionEvent("<@W23456789> foo"), smClient)

	post := <-ts.post
	assert.Equal(t, `foo`, post.Get("text"))

	// invalid command
	go b.SlackApp.SocketModeHandler.(*testutils.FakeHandler).SendEvent(testutils.AppMentionEvent("<@W23456789> bar"), smClient)

	post = <-ts.post
	assert.Equal(t, `[{"color":"bad","title":"invalid command","text":"supported commands: foo","blocks":null}]`, post.Get("attachments"))

	cancel()
	assert.NoError(t, <-errCh)
}

func Test_tokenizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single word",
			input: "foo",
			want:  []string{"foo"},
		},
		{
			name:  "multiple words",
			input: "foo bar",
			want:  []string{"foo", "bar"},
		},
		{
			name:  "quoted phrase",
			input: `foo "bar snafu" foo`,
			want:  []string{"foo", "bar snafu", "foo"},
		},
		{
			name:  "mismatched quotes",
			input: `foo "bar snafu`,
			want:  []string{"foo", "bar", "snafu"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tokenizeText(tt.input))
		})
	}
}

func Test_removeUserID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no mention",
			input: "foo",
			want:  "foo",
		},
		{
			name:  "match",
			input: "<@U07V31R90R0> foo",
			want:  "foo",
		},
		{
			name:  "match empty",
			input: "<@U07V31R90R0> ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, removeUserID(tt.input))
		})
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type testServer struct {
	t    *testing.T
	post chan url.Values
}

func (s *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth.test":
		_, _ = w.Write([]byte(`{ "ok": true, "url": "https://subarachnoid.slack.com/", "team": "Subarachnoid Workspace", "user": "bot", "team_id": "T0G9PQBBK", "user_id": "W23456789", "bot_id": "BZYBOTHED" }`))
	case "/chat.postMessage":
		body, _ := io.ReadAll(r.Body)
		if values, err := url.ParseQuery(string(body)); err == nil {
			s.post <- values
		}
	default:
		s.t.Log(r.URL.String())
		http.Error(w, "not found", http.StatusNotFound)
	}
}
