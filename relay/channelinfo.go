package relay

import "github.com/nlopes/slack"

type ChannelInfo struct {
	ID      string
	Members []string
	Name    string
}

func newChannelInfoFromSlackChannel(slackChannel *slack.Channel) *ChannelInfo {
	return (&ChannelInfo{}).setFromSlackChennel(slackChannel)
}

func (ch *ChannelInfo) setFromSlackChennel(slackChannel *slack.Channel) *ChannelInfo {
	ch.ID = slackChannel.ID
	ch.Members = slackChannel.Members
	ch.Name = slackChannel.Name

	return ch
}
