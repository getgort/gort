package adapter

import (
	"github.com/slack-go/slack"
)

// UserInfo contains the basic information for a single user in any chat provider.
type UserInfo struct {
	ID                    string
	Name                  string
	DisplayName           string
	DisplayNameNormalized string
	Email                 string
	FirstName             string
	LastName              string
	RealName              string
	RealNameNormalized    string
}

func newUserInfoFromSlackUser(slackUser *slack.User) *UserInfo {
	return (&UserInfo{}).setFromSlackUser(slackUser)
}

func (u *UserInfo) setFromSlackUser(slackUser *slack.User) *UserInfo {
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
