package slackapp

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log/slog"
	"sync/atomic"
)

type SlackApp struct {
	*socketmode.Client
	Events chan slackevents.EventsAPIInnerEvent
	SocketModeHandler
	logger    *slog.Logger
	connected atomic.Bool
}

type SocketModeHandler interface {
	RunEventLoopContext(ctx context.Context) error
	Handle(socketmode.EventType, socketmode.SocketmodeHandlerFunc)
}

func NewSlackApp(client *slack.Client, logger *slog.Logger) *SlackApp {
	smc := socketmode.New(client)
	return newSlackAppWithSocketModeHandler(smc, socketmode.NewSocketmodeHandler(smc), logger)
}

func newSlackAppWithSocketModeHandler(client *socketmode.Client, handler SocketModeHandler, logger *slog.Logger) *SlackApp {
	app := SlackApp{
		Client:            client,
		Events:            make(chan slackevents.EventsAPIInnerEvent),
		SocketModeHandler: handler,
		logger:            logger,
	}
	app.SocketModeHandler.Handle(socketmode.EventTypeConnecting, app.onConnecting)
	app.SocketModeHandler.Handle(socketmode.EventTypeConnectionError, app.onConnectionError)
	app.SocketModeHandler.Handle(socketmode.EventTypeConnected, app.onConnected)
	app.SocketModeHandler.Handle(socketmode.EventTypeIncomingError, app.onIncomingError)
	app.SocketModeHandler.Handle(socketmode.EventTypeHello, app.onHello)
	app.SocketModeHandler.Handle(socketmode.EventTypeDisconnect, app.onDisconnected)
	app.SocketModeHandler.Handle(socketmode.EventTypeEventsAPI, app.onEvent)

	return &app
}

func (h *SlackApp) Run(ctx context.Context) error {
	h.logger.Info("starting SlackApp")
	defer h.logger.Info("shutting down SlackApp")
	return h.SocketModeHandler.RunEventLoopContext(ctx)
}

func (h *SlackApp) Connected() bool {
	return h.connected.Load()
}

func (h *SlackApp) onHello(_ *socketmode.Event, _ *socketmode.Client) {
}

func (h *SlackApp) onConnecting(_ *socketmode.Event, _ *socketmode.Client) {
	h.logger.Debug("connecting to Slack ...")
}

func (h *SlackApp) onConnectionError(evt *socketmode.Event, _ *socketmode.Client) {
	reason := string(evt.Type)
	if evt.Request != nil {
		reason = evt.Request.Reason
	}
	h.logger.Error("failed to connect to Slack", "reason", reason)
}

func (h *SlackApp) onConnected(_ *socketmode.Event, _ *socketmode.Client) {
	h.connected.Store(true)
	h.logger.Info("connected to Slack")
}

func (h *SlackApp) onDisconnected(_ *socketmode.Event, _ *socketmode.Client) {
	h.connected.Store(false)
	h.logger.Warn("disconnected from Slack")
}

func (h *SlackApp) onEvent(evt *socketmode.Event, client *socketmode.Client) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		h.logger.Warn("received unexpected event type", "type", evt.Type)
		return
	}
	client.Ack(*evt.Request)
	innerEvent := eventsAPIEvent.InnerEvent
	h.logger.Debug("Event received", "type", innerEvent.Type)

	h.Events <- innerEvent
}

func (h *SlackApp) onIncomingError(event *socketmode.Event, _ *socketmode.Client) {
	// TODO
	h.logger.Warn("received unexpected event", "type", event.Type, "type", fmt.Sprintf("%T", event.Data))
}