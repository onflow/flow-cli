/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

const (
	MIXPANEL_EVENT_PROJECT_TOKEN = "token"
	MIXPANEL_EVENT_CALLER        = "caller"

	FLOW_CLI               = "flow-cli"
	MIXPANEL_PROJECT_TOKEN = "da53727e0435e12820c831098691f8e5"
)

type Event struct {
	Name       string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
}

func NewEvent(name string) *Event {
	return &Event{
		name,
		make(map[string]interface{}),
	}
}
func (e *Event) SetUpEvent(token string, caller string) {
	e.Properties[MIXPANEL_EVENT_PROJECT_TOKEN] = token
	e.Properties[MIXPANEL_EVENT_CALLER] = caller
}
