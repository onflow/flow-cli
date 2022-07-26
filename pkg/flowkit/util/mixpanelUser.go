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

import (
	"encoding/base64"
	"os/user"
	"strings"
)

type mixpanelUser struct {
	Token      string                 `json:"$token"`
	DistinctId string                 `json:"$distinct_id"`
	Set        map[string]interface{} `json:"$set"`
}

func getMixPanelUser() (*mixpanelUser, error) {
	distinctId, err := generateNewDistinctId()
	if err != nil {
		return nil, err
	}
	return &mixpanelUser{
		MIXPANEL_PROJECT_TOKEN,
		distinctId,
		make(map[string]interface{}),
	}, nil
}

func (e *mixpanelUser) disableUserTracking() {
	e.Set["opt_in"] = false
}

func generateNewDistinctId() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	name := currentUser.Name
	editedName := strings.Replace(name, " ", "-", -1)
	username := currentUser.Username
	id := currentUser.Uid

	combinedString := editedName + username + id
	encodedString := base64.StdEncoding.EncodeToString([]byte(combinedString))

	return encodedString, nil
}
