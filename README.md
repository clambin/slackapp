# slackapp
[![release](https://img.shields.io/github/v/tag/clambin/slackapp?color=green&label=release&style=plastic)](https://github.com/clambin/slackapp/releases)
[![codecov](https://img.shields.io/codecov/c/gh/clambin/slackapp?style=plastic)](https://app.codecov.io/gh/clambin/slackapp)
[![test](https://github.com/clambin/slackapp/workflows/test/badge.svg)](https://github.com/clambin/slackapp/actions)
[![go report card](https://goreportcard.com/badge/github.com/clambin/slackapp)](https://goreportcard.com/report/github.com/clambin/slackapp)
[![godoc](https://pkg.go.dev/badge/github.com/clambin/slackapp?utm_source=godoc)](https://pkg.go.dev/github.com/clambin/slackapp)
[![license](https://img.shields.io/github/license/clambin/slackapp?style=plastic)](LICENSE.md)

A basic Slack Events API client for Go, using [github.com/go-slack/slack](http://github.com/slack-go/slack)

## Overview

This packages provides a Go implementation of Slack's [Events API], a streamlined way to build apps and bots that respond to activities in Slack.

The slackapp implementation uses Slack's [Socket Mode] to establish the connection, meaning apps do not need
to designate a public HTTP endpoint for Slack to connect to.

[Events API]: https://api.slack.com/apis/events-api
[Socket Mode]: https://api.slack.com/apis/socket-mode

## Installing

go get

```
$ go get -u github.com/slack-go/slack
```

## Slack configuration

To create a SlackApp, you will need to add it to your workspace. In short, this means:

- Go to [Your Apps](https://api.slack.com/apps) in your workspace and "Create a New App".
- In "Socket Mode", enable Socket Mode.
- In "App Home", note down your App-Level Token (it starts with `xapp-1`)
- In "OAuth & Permissions", set the required permissions (see [manifest.yaml](assets/slack/manifest.yaml) for a basic example) and note your Bot User OAuth Token token (it starts with `xoxb-`).
- In "Events Subscriptions", enable events and subscribe to the bot events needed for your application (e.g. for a Bot, you will at least need `app_mention`).

## Writing a SlackApp client

See [doc_slackapp_test.go](doc_slackapp_test.go) for a basic example of a SlackApp client.

## Bot

Additionally, this module contains a basic implementation of a Events API-based Slack Bot. It connects to Slack and waits 
for it to be mentioned by the user.  It then executes the command, and posts the output to the channel.

See [doc_bot_test.go](doc_bot_test.go) for an example of a Bot.

## Authors

* **Christophe Lambin**

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
