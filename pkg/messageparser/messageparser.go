// Package messageparser provides mechanics to parse a slack channel
// history for messages containing the information we are looking for,
// which is alert/incident context.
package messageparser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
)

type ParseResult struct {
	IsAlert                  bool   `json:"-"`
	IsDone                   bool   `json:"-"`
	AlertName                string `json:"alert_name,omitempty"`
	Priority                 string `json:"priority,omitempty"`
	InstallationName         string `json:"installation_name,omitempty"`
	InstallationPipeline     string `json:"installation_pipeline,omitempty"`
	Provider                 string `json:"provider,omitempty"`
	AffectsManagementCluster bool   `json:"affects_management_cluster"`
	AffectsWorkloadCluster   bool   `json:"affects_workload_cluster"`
	WorkloadClusterID        string `json:"workload_cluster_id,omitempty"`
	SlackChannelID           string `json:"slack_channel_id"`
	SlackChannelName         string `json:"slack_channel_name"`
}

var titleRegex = regexp.MustCompile(`#[0-9]+: \[([A-Za-z0-9]+)\]: ([a-z]+) / ([a-z0-9]+) - (.+)`)

func ParseConversationHistory(history *slack.GetConversationHistoryResponse, doneString string) (ParseResult, error) {
	for _, msg := range history.Messages {
		res, err := ParseMessage(msg, doneString)
		if err != nil {
			fmt.Printf("ERROR: Message could not be parsed. %s", err)
		}

		if res.IsAlert {
			return res, nil
		}
	}

	// Found nothing. Return empty result.
	return ParseResult{}, nil
}

func ParseMessage(message slack.Message, doneString string) (ParseResult, error) {
	result := ParseResult{}

	// Find done string
	if strings.Contains(message.Msg.Text, doneString) {
		result.IsDone = true
		return result, nil
	}

	if len(message.Msg.Attachments) == 0 {
		return result, nil
	}

	for _, attachment := range message.Msg.Attachments {
		// Typical opsgenie properties
		if strings.Contains(attachment.TitleLink, "https://opsg.in/") {
			result.IsAlert = true

			// Parse typical title
			titleParts := titleRegex.FindStringSubmatch(attachment.Title)
			fmt.Printf("Title parts: %#v\n", titleParts)
			if len(titleParts) == 5 {
				result.InstallationName = titleParts[2]
				result.AlertName = titleParts[4]
			}

			for _, field := range attachment.Fields {
				if field.Title == "Priority" {
					result.Priority = field.Value

				} else if field.Title == "Tags" {
					// Tags are completely unstructured. So our parsing is dull, too.
					tags := getTags(field.Value)

					// Workload cluster/management cluster
					if sliceContains(tags, "management_cluster") {
						result.AffectsManagementCluster = true
					}
					if sliceContains(tags, "workload_cluster") {
						result.AffectsWorkloadCluster = true
						if len(titleParts) == 5 {
							result.WorkloadClusterID = titleParts[3]
						}
					}

					if sliceContains(tags, "aws") {
						result.Provider = "aws"
					} else if sliceContains(tags, "azure") {
						result.Provider = "azure"
					} else if sliceContains(tags, "kvm") {
						result.Provider = "kvm"
					}

					if sliceContains(tags, "stable") {
						result.InstallationPipeline = "stable"
					} else if sliceContains(tags, "testing") {
						result.InstallationPipeline = "testing"
					}
				}
			}
		}
	}

	return result, nil
}

// getTags splits a string by comma and makes sure there is no whitespaces left.
func getTags(longString string) []string {
	return regexp.MustCompile(`,\s*`).Split(longString, -1)
}

func sliceContains(haystack []string, needle string) bool {
	for _, piece := range haystack {
		if piece == needle {
			return true
		}
	}
	return false
}
