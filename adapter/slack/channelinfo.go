package slack

import (
	"github.com/clockworksoul/cog2/adapter"
	"github.com/slack-go/slack"
)

func newChannelInfoFromSlackChannel(slackChannel *slack.Channel) *adapter.ChannelInfo {
	ch := &adapter.ChannelInfo{}

	ch.ID = slackChannel.ID
	ch.Members = slackChannel.Members
	ch.Name = slackChannel.Name

	return ch
}
