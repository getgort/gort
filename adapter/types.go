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

package adapter

// MessageRef is a way to refer to a unique message across all adapters.
// Different fields are non-zero based on what variety of adapter is referred
// to. This is only given to bundle commands as a json-encoded string.
type MessageRef struct {
	// ID is the id of the message.
	// Used by: discord.
	ID string

	// ChannelID is the id of the channel containing the message.
	// Used by: discord, slack.
	ChannelID string

	// Timestamp is the time the message was sent.
	// Used by: slack.
	Timestamp string

	// Adapter is the name of the adapter the message is in.
	Adapter string
}
