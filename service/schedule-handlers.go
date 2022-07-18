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
	"fmt"
	"net/http"
	"strconv"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/scheduler"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// handleScheduleCommand handles "PUT /v2/schedules"
func handleScheduleCommand(w http.ResponseWriter, r *http.Request) {
	var scheduleRequest rest.ScheduleRequest
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

	id, err := scheduler.ScheduleFromString(r.Context(), scheduleRequest.Command, sc)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	_, err = fmt.Fprintf(w, "%d", id)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handleGetSchedules handles "GET /v2/schedules"
func handleGetSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := scheduler.GetSchedules(r.Context())
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	info := make([]rest.ScheduleInfo, 0)

	for _, s := range schedules {
		a, err := adapter.GetAdapter(s.Adapter)
		if err != nil {
			respondAndLogError(r.Context(), w, err)
		}
		ch, err := a.GetChannelInfo(s.ChannelID)
		if err != nil {
			respondAndLogError(r.Context(), w, err)
		}
		i := rest.ScheduleInfo{
			Id:          s.ScheduleID,
			Command:     s.Command.Original,
			Cron:        s.Cron,
			Adapter:     s.Adapter,
			ChannelName: ch.Name,
		}

		info = append(info, i)
	}

	json.NewEncoder(w).Encode(info)
}

func handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	err = scheduler.Cancel(r.Context(), id)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

func addScheduleMethodsToRouter(router *mux.Router) {
	router.Handle("/v2/schedules", otelhttp.NewHandler(authCommand(handleScheduleCommand, "schedule"), "handleScheduleCommand")).Methods("PUT")
	router.Handle("/v2/schedules", otelhttp.NewHandler(authCommand(handleGetSchedules, "schedule", "get"), "handleGetSchedules")).Methods("GET")
	router.Handle("/v2/schedules/{id:\\d+}", otelhttp.NewHandler(authCommand(handleDeleteSchedule, "schedule", "delete"), "handleDeleteSchedule")).Methods("DELETE")
}
