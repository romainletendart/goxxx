// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

// Help package
package help

import (
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"strings"
)

var messageList []string

// Store messages to display them later via the help command
func AddMessages(helpMessages ...string) {
	for _, message := range helpMessages {
		messageList = append(messageList, message)
	}
}

// Handler for the !help command
func HandleHelpCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	if len(fields) == 0 || fields[0] != "!help" {
		return false
	}
	for _, message := range messageList {
		callback(&core.ReplyCallbackData{Message: message, Nick: event.Nick})
	}
	return true
}
