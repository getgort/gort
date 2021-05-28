/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
