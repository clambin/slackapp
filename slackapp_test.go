package slackapp

import (
	"context"
	"errors"
	"github.com/clambin/slackapp/internal/testutils"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

func TestSlackApp(t *testing.T) {
	var h testutils.FakeHandler
	app := newSlackAppWithSocketModeHandler(nil, &h, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errChan := make(chan error)
	go func() { errChan <- app.Run(ctx) }()

	// connect
	assert.False(t, app.Connected())
	app.onConnecting(nil, nil)
	app.onConnected(nil, nil)
	assert.True(t, app.Connected())
	app.onHello(nil, nil)

	// invalid events are ignored
	go app.onEvent(&socketmode.Event{Type: socketmode.EventTypeInvalidAuth}, nil)
	assert.Empty(t, app.Events)

	// valid events are sent to app.Events
	slackClient := slack.New("", slack.OptionHTTPClient(&http.Client{Transport: &testutils.StubbedRoundTripper{}}))
	go app.onEvent(testutils.AppMentionEvent("hello world"), socketmode.New(slackClient))
	evt := <-app.Events
	assert.Equal(t, string(slackevents.AppMention), evt.Type)
	mention, ok := evt.Data.(*slackevents.AppMentionEvent)
	require.True(t, ok)
	assert.Equal(t, "hello world", mention.Text)

	// connection error / disconnect
	app.onIncomingError(&socketmode.Event{Data: &slack.IncomingEventError{ErrorObj: errors.New("fail")}}, nil)
	ev := socketmode.Event{
		Type:    socketmode.EventTypeInvalidAuth,
		Request: &socketmode.Request{Reason: "invalid credentials"},
	}
	app.onConnectionError(&ev, nil)
	app.onDisconnected(nil, nil)
	assert.False(t, app.Connected())

	// shutdown
	cancel()
	assert.NoError(t, <-errChan)
}
