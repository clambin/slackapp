package testutils

import (
	"bytes"
	"context"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"io"
	"net/http"
)

type FakeHandler struct {
	eventHandlers map[socketmode.EventType]socketmode.SocketmodeHandlerFunc
}

func (f *FakeHandler) RunEventLoopContext(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (f *FakeHandler) Handle(evt socketmode.EventType, h socketmode.SocketmodeHandlerFunc) {
	if f.eventHandlers == nil {
		f.eventHandlers = make(map[socketmode.EventType]socketmode.SocketmodeHandlerFunc)
	}
	f.eventHandlers[evt] = h
}

func (f *FakeHandler) SendEvent(ev *socketmode.Event, c *socketmode.Client) {
	if h, ok := f.eventHandlers[socketmode.EventTypeEventsAPI]; ok {
		h(ev, c)
	}
}

func (f *FakeHandler) Login() {
	f.eventHandlers[socketmode.EventTypeConnected](nil, nil)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ http.RoundTripper = &StubbedRoundTripper{}

type StubbedRoundTripper struct{}

func (r StubbedRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(``)),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func AppMentionEvent(text string) *socketmode.Event {
	return &socketmode.Event{
		Request: &socketmode.Request{},
		Data: slackevents.EventsAPIEvent{
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Type: string(slackevents.AppMention),
				Data: &slackevents.AppMentionEvent{
					Channel: "1",
					Text:    text,
				},
			},
		},
	}
}
