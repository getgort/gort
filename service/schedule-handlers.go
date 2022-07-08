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

package service

import (
	"encoding/json"
	"net/http"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/scheduler"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ScheduleRequest struct {
	Command   string
	Cron      string
	Adapter   string
	ChannelID string
}

func handleScheduleCommand(w http.ResponseWriter, r *http.Request) {
	var scheduleRequest ScheduleRequest
	err := json.NewDecoder(r.Body).Decode(&scheduleRequest)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	user, err := getUserByRequest(r)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	userID := user.Mappings[scheduleRequest.Adapter]

	sc := data.ScheduledCommand{
		Adapter:   scheduleRequest.Adapter,
		ChannelID: scheduleRequest.ChannelID,
		UserID:    userID,
		UserEmail: user.Email,
		UserName:  user.Username,
		Cron:      scheduleRequest.Cron,
	}

	err = scheduler.AddFromString(r.Context(), scheduleRequest.Command, sc)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

func addScheduleMethodsToRouter(router *mux.Router) {
	router.Handle("/v2/schedule/create", otelhttp.NewHandler(authCommand(handleScheduleCommand, "schedule"), "handleScheduleCommand")).Methods("POST")
}
