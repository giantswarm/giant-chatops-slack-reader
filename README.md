[![CircleCI](https://circleci.com/gh/giantswarm/giant-chatops-slack-reader.svg?style=shield)](https://circleci.com/gh/giantswarm/giant-chatops-slack-reader)

# giant-chatops Slack Reader

This little utility is triggered with a Slack channel ID and a Slack token,
to

- go and look for alert details
- pass these details on for use in Argo Workflows
- post some status message back into the channel

Required environment variables:

- `SLACK_TOKEN`: The Slack auth token
- `SLACK_CHANNEL_ID`: ID of the Slack channel this program should use
