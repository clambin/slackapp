package slackapp

import (
	"context"
	"errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log/slog"
	"sync/atomic"
)

// A SlackApp implements Slack's Events API, using Socket Mode. It connects to Slack,  listens for incoming events
// and makes them available using the Event channel.
type SlackApp struct {
	*socketmode.Client
	Events chan slackevents.EventsAPIInnerEvent
	socketModeHandler
	logger    *slog.Logger
	connected atomic.Bool
}

type socketModeHandler interface {
	RunEventLoopContext(ctx context.Context) error
	Handle(socketmode.EventType, socketmode.SocketmodeHandlerFunc)
}

// NewSlackApp creates a new slackapp for the slack client.
func NewSlackApp(client *slack.Client, logger *slog.Logger) *SlackApp {
	smc := socketmode.New(client)
	return newSlackAppWithSocketModeHandler(smc, socketmode.NewSocketmodeHandler(smc), logger)
}

func newSlackAppWithSocketModeHandler(client *socketmode.Client, handler socketModeHandler, logger *slog.Logger) *SlackApp {
	app := SlackApp{
		Client:            client,
		Events:            make(chan slackevents.EventsAPIInnerEvent),
		socketModeHandler: handler,
		logger:            logger,
	}
	app.socketModeHandler.Handle(socketmode.EventTypeConnecting, app.onConnecting)
	app.socketModeHandler.Handle(socketmode.EventTypeConnectionError, app.onConnectionError)
	app.socketModeHandler.Handle(socketmode.EventTypeConnected, app.onConnected)
	app.socketModeHandler.Handle(socketmode.EventTypeIncomingError, app.onIncomingError)
	app.socketModeHandler.Handle(socketmode.EventTypeHello, app.onHello)
	app.socketModeHandler.Handle(socketmode.EventTypeDisconnect, app.onDisconnected)
	app.socketModeHandler.Handle(socketmode.EventTypeEventsAPI, app.onEvent)

	return &app
}

// Run starts the slackapp. It connects to Slack and passes any received events to the Events channel.
func (h *SlackApp) Run(ctx context.Context) error {
	h.logger.Info("starting SlackApp")
	defer h.logger.Info("shutting down SlackApp")
	return h.socketModeHandler.RunEventLoopContext(ctx)
}

// Connected returns true if the slackapp is connected to Slack.
func (h *SlackApp) Connected() bool {
	return h.connected.Load()
}

func (h *SlackApp) onConnecting(_ *socketmode.Event, _ *socketmode.Client) {
	h.logger.Debug("connecting to Slack ...")
}

func (h *SlackApp) onConnectionError(ev *socketmode.Event, _ *socketmode.Client) {
	reason := string(ev.Type)
	if ev.Request != nil {
		reason = ev.Request.Reason
	}
	h.logger.Error("failed to connect to Slack", "reason", reason)
}

func (h *SlackApp) onConnected(_ *socketmode.Event, _ *socketmode.Client) {
	h.connected.Store(true)
	h.logger.Info("connected to Slack")
}

func (h *SlackApp) onIncomingError(ev *socketmode.Event, _ *socketmode.Client) {
	var err *slack.IncomingEventError
	if errors.As(ev.Data.(error), &err) {
		h.logger.Warn("received incoming error", "err", err)
	} else {
		h.logger.Warn("received unexpected event type", "type", ev.Type)
	}
}

func (h *SlackApp) onHello(_ *socketmode.Event, _ *socketmode.Client) {
}

func (h *SlackApp) onDisconnected(_ *socketmode.Event, _ *socketmode.Client) {
	h.connected.Store(false)
	h.logger.Warn("disconnected from Slack")
}

func (h *SlackApp) onEvent(ev *socketmode.Event, client *socketmode.Client) {
	eventsAPIEvent, ok := ev.Data.(slackevents.EventsAPIEvent)
	if !ok {
		h.logger.Warn("received unexpected event type", "type", ev.Type)
		return
	}
	client.Ack(*ev.Request)
	innerEvent := eventsAPIEvent.InnerEvent
	h.logger.Debug("Event received", "type", innerEvent.Type)

	h.Events <- innerEvent
}
