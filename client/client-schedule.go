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

package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getgort/gort/data/rest"
)

// ScheduleCreate schedules a new command to run periodically.
func (c *GortClient) ScheduleCreate(req rest.ScheduleRequest) error {
	url := fmt.Sprintf("%s/v2/schedule", c.profile.URL.String())

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("PUT", url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getResponseError(resp)
	}

	return nil
}
