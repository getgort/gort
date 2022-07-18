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

package rest

// ScheduleRequest holds all the information needed by the Gort server to
// schedule a recurring command.
type ScheduleRequest struct {
	// Command is the full command to be run every time.
	Command string

	// Cron is the specification of when the command should be run, in standard
	// cron format.
	Cron string

	// Adapter is the name of the adapter that the command should output to.
	Adapter string

	// ChannelID is the id of the channel that the command should output to.
	ChannelID string
}

type ScheduleInfo struct {
	Id          int64
	Command     string
	Cron        string
	Adapter     string
	ChannelName string
}
