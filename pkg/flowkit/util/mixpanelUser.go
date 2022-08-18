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
	"crypto/sha256"
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
	distinctId, err := GenerateNewDistinctId()
	if err != nil {
		return nil, err
	}
	return &mixpanelUser{
		MIXPANEL_PROJECT_TOKEN,
		distinctId,
		make(map[string]interface{}),
	}, nil
}

func (e *mixpanelUser) configureUserTracking(enable bool) {
	e.Set["opt_in"] = enable
	e.Set["$city"] = nil
	e.Set["$region"] = nil
	e.Set["$country_code"] = nil
	e.Set["$geo_source"] = nil
	e.Set["$timezone"] = nil
}

func GenerateNewDistinctId() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	name := currentUser.Name
	hyphenatedName := strings.Replace(name, " ", "-", -1)
	username := currentUser.Username
	id := currentUser.Uid

	combinedString := hyphenatedName + username + id

	hashedString := sha256.Sum256([]byte(combinedString))
	encodedString := base64.StdEncoding.EncodeToString(hashedString[:])

	return encodedString, nil
}
