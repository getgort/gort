package slack

import (
	"github.com/clockworksoul/gort/adapter"
	"github.com/slack-go/slack"
)

func newUserInfoFromSlackUser(slackUser *slack.User) *adapter.UserInfo {
	u := &adapter.UserInfo{}

	u.ID = slackUser.ID
	u.Name = slackUser.Name
	u.DisplayName = slackUser.Profile.DisplayName
	u.DisplayNameNormalized = slackUser.Profile.DisplayNameNormalized
	u.Email = slackUser.Profile.Email
	u.FirstName = slackUser.Profile.FirstName
	u.LastName = slackUser.Profile.LastName
	u.RealName = slackUser.RealName
	u.RealNameNormalized = slackUser.Profile.RealNameNormalized

	return u
}
