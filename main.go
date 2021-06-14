// Package main is a little Slack client kumping into a freshly created channel, looking
// for messages containing information about an incident.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"

	"github.com/giantswarm/giant-chatops-slack-reader/pkg/messageparser"
)

const (
	argoWebhookURL = "http://giant-chatops-alert-eventsource-svc:12000/alert"
)

func main() {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		fmt.Println("ERROR: Environment variable SLACK_TOKEN must be set.")
		os.Exit(0)
	}

	slackChannelID := os.Getenv("SLACK_CHANNEL_ID")
	if slackChannelID == "" {
		fmt.Println("ERROR: Environment variable SLACK_CHANNEL_ID must be set.")
		os.Exit(0)
	}

	fmt.Printf("Channel ID: %s\n", slackChannelID)

	api := slack.New(slackToken, slack.OptionDebug(true))

	// Find out channel name
	channel, err := api.GetConversationInfo(slackChannelID, false)
	if err != nil {
		fmt.Printf("ERROR: Cannot get channel details - %#v\n", err)
		os.Exit(0)
	}
	fmt.Printf("Channel name: %#v\n", channel.Name)
	if !strings.HasPrefix(channel.Name, "inc-") {
		fmt.Println("Ignoring this channel, as the name does not start with 'inc-'. Exiting.\n", channel.Name)
		os.Exit(0)
	}

	// Simply wait for 5 seconds, so there is enough time for the alert details to appear.
	time.Sleep(5 * time.Second)

	_, _, _, err = api.JoinConversation(slackChannelID)
	if err != nil {
		fmt.Printf("ERROR: Could not join channel %s: %#v\n", slackChannelID, err)
		os.Exit(0)
	}

	// We set this to prevent processing the same channel multiple times.
	doneIdentifier := fmt.Sprintf("[giant-chatops-slack-reader done for channel %s]", slackChannelID)

	history, err := api.GetConversationHistory(&slack.GetConversationHistoryParameters{ChannelID: slackChannelID})
	if err != nil {
		fmt.Printf("ERROR: Could not get conversation history: %#v\n", err)
		os.Exit(0)
	}

	parseResult, err := messageparser.ParseConversationHistory(history, doneIdentifier)
	if err != nil {
		fmt.Printf("ERROR: Could not parse conversation history: %#v\n", err)
		os.Exit(0)
	}

	parseResult.SlackChannelID = slackChannelID
	parseResult.SlackChannelName = channel.Name

	// Trigger event via webhook. Collect any errors in eventError.
	var eventError error
	if parseResult.IsAlert {
		jsonBytes, err := json.Marshal(parseResult)
		if err == nil {
			resp, err := http.Post(argoWebhookURL, "application/json", bytes.NewBuffer(jsonBytes))
			if err != nil {
				eventError = err
			} else if resp.StatusCode != http.StatusOK {
				eventError = fmt.Errorf("could not trigger alert event, webhook status was %d", resp.StatusCode)
			}
		} else {
			eventError = err
		}
	}

	// Give feedback to channel
	text := ""
	if eventError != nil {
		text = fmt.Sprintf("Could not trigger alert event: `%s`", eventError)
	} else {
		if parseResult.IsAlert {
			text = "I'm passing these alert details to my co-bots to gather some information for you:\n\n"
			jsonBytes, err := json.MarshalIndent(parseResult, "", "    ")
			if err != nil {
				text += fmt.Sprintf("Problem writing JSON: `%s`", err)
			} else {
				text += fmt.Sprintf("```%s```\n", string(jsonBytes))
			}

			text += fmt.Sprintf("\n\n`%s`", doneIdentifier)
		} else {
			text = "No alert info found so far. Please share an #opsgenie alert message in this channel."
		}
	}

	_, _, _, err = api.SendMessage(slackChannelID, slack.MsgOptionText(text, false))
	if err != nil {
		fmt.Printf("ERROR: Could not post feedback: %#v\n", err)
	}
}
