// Package messageparser provides mechanics to parse a slack channel
// history for messages containing the information we are looking for,
// which is alert/incident context.
package messageparser

import (
	"reflect"
	"testing"

	"github.com/slack-go/slack"
)

func TestParseMessage(t *testing.T) {
	type args struct {
		message    slack.Message
		doneString string
	}
	tests := []struct {
		name    string
		args    args
		want    ParseResult
		wantErr bool
	}{
		{
			name: "Simple non-alert message",
			args: args{
				message: slack.Message{
					Msg: slack.Msg{
						ClientMsgID: "d2d227c0-a1ce-46a4-85b0-952a5b16f5a1",
						Type:        "message",
						Channel:     "",
						User:        "U02MCG949",
						Text:        "Test message",
						Timestamp:   "1623338820.000600",
						Attachments: []slack.Attachment(nil),
						Edited:      (*slack.Edited)(nil),
						Icons:       (*slack.Icon)(nil),
						BotProfile:  (*slack.BotProfile)(nil),
						Members:     []string(nil),
						Replies:     []slack.Reply(nil),
						Files:       []slack.File(nil),
						Comment:     (*slack.Comment)(nil),
						Team:        "T0251EQJH",
						Reactions:   []slack.ItemReaction(nil),
						Blocks:      slack.Blocks{BlockSet: []slack.Block{nil}},
					},
					SubMessage:      (*slack.Msg)(nil),
					PreviousMessage: (*slack.Msg)(nil),
				},
				doneString: "[done]",
			},
			want:    ParseResult{},
			wantErr: false,
		},
		{
			name: "Message with done string",
			args: args{
				message: slack.Message{
					Msg: slack.Msg{
						ClientMsgID: "d2d227c0-a1ce-46a4-85b0-952a5b16f5a1",
						Type:        "message",
						Channel:     "",
						User:        "U02MCG949",
						Text:        "Test message [done]",
						Timestamp:   "1623338820.000600",
						Attachments: []slack.Attachment(nil),
						Edited:      (*slack.Edited)(nil),
						Icons:       (*slack.Icon)(nil),
						BotProfile:  (*slack.BotProfile)(nil),
						Members:     []string(nil),
						Replies:     []slack.Reply(nil),
						Files:       []slack.File(nil),
						Comment:     (*slack.Comment)(nil),
						Team:        "T0251EQJH",
						Reactions:   []slack.ItemReaction(nil),
						Blocks:      slack.Blocks{BlockSet: []slack.Block{nil}},
					},
					SubMessage:      (*slack.Msg)(nil),
					PreviousMessage: (*slack.Msg)(nil),
				},
				doneString: "[done]",
			},
			want: ParseResult{
				IsDone: true,
			},
			wantErr: false,
		},
		{
			name: "Opsgenie alert message shared in channel",
			args: args{
				message: slack.Message{
					Msg: slack.Msg{
						Type:      "message",
						User:      "U02MCG949",
						Timestamp: "1623339149.000700",
						PinnedTo:  []string(nil),
						Attachments: []slack.Attachment{
							{
								Color:      "D0D0D0",
								Fallback:   "\"[Prometheus]: anteater / anteater - PrometheusPersistentVolumeSpaceTooLow\" <https://opsg.in/a/i/giantswarm/2e674f86-b131-42be-9f9c-3c68b49684af-1623336728751|4865>\nTags: PrometheusPersistentVolumeSpaceTooLow, anteater, atlas, aws, management_cluster, page, stable",
								CallbackID: "714426f1-1d58-46a3-b9d1-35d6adb7f9d3_83721432-4262-4b27-8420-6e11cefe5023_4865",
								ID:         1,
								Title:      "#4865: [Prometheus]: anteater / anteater - PrometheusPersistentVolumeSpaceTooLow",
								TitleLink:  "https://opsg.in/a/i/giantswarm/2e674f86-b131-42be-9f9c-3c68b49684af-1623336728751",
								Text:       "*Team:* atlas\n*Area:* empowerment / observability\n*Recipe:* <https://intranet.giantswarm.io/docs/support-and-ops/ops-recipes/low-disk-space/#persistent-volume>\n :fire: 172.19.5.231:10300: Persistent volume /var/lib/kubelet/pods/25526e0c-b131-48cd-aef3-392637c259b5/volume-subpaths/pvc-eb616885-db34-40c5-9349-cb5aa6363a1f/prometheus/2 on 172.19.5.231:10300 does not have enough free space.\n :fire: 172.19.5.231:10300: Persistent volume /var/lib/kubelet/pods/25526e0c-b131-48cd-aef3-392637c259b5/volumes/kubernetes.io~aws-ebs/pvc-eb616885-db34-40c5-9349-cb5aa6363a1f on 172.19.5.231:10300 does not have enough free space.",
								FromURL:    "https://gigantic.slack.com/archives/C09TB4L5C/p1623336729052200",
								Fields: []slack.AttachmentField{
									{
										Title: "Priority",
										Value: "P3",
										Short: true,
									},
									{
										Title: "Tags",
										Value: "PrometheusPersistentVolumeSpaceTooLow, anteater, atlas, aws, management_cluster, page, stable",
										Short: true,
									},
									{
										Title: "Routed Teams",
										Value: "atlas, alerts_router_team",
										Short: true,
									},
								},
								Actions:    []slack.AttachmentAction(nil),
								MarkdownIn: []string{"text", "text"},
								Blocks:     slack.Blocks{BlockSet: []slack.Block(nil)},
							},
						},
						Edited:     (*slack.Edited)(nil),
						Icons:      (*slack.Icon)(nil),
						BotProfile: (*slack.BotProfile)(nil),
						Members:    []string(nil),
						Replies:    []slack.Reply(nil),
						Files:      []slack.File(nil),
						Comment:    (*slack.Comment)(nil),
						Team:       "T0251EQJH",
						Reactions:  []slack.ItemReaction(nil),
						Blocks:     slack.Blocks{BlockSet: []slack.Block(nil)}},
					SubMessage:      (*slack.Msg)(nil),
					PreviousMessage: (*slack.Msg)(nil),
				},
				doneString: "[done]",
			},
			want: ParseResult{
				IsAlert:                  true,
				AlertName:                "PrometheusPersistentVolumeSpaceTooLow",
				Priority:                 "P3",
				InstallationName:         "anteater",
				InstallationPipeline:     "stable",
				Provider:                 "aws",
				AffectsManagementCluster: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMessage(tt.args.message, tt.args.doneString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTags(t *testing.T) {
	type args struct {
		longString string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "First",
			args: args{longString: "foo, bar, baz"},
			want: []string{"foo", "bar", "baz"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTags(tt.args.longString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
