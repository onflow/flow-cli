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
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/user"
	"strings"

	"github.com/spf13/cobra"
)

const (
	MIXPANEL_TRACK_URL = "https://api.mixpanel.com/track"
)

var MIXPANEL_PROJECT_TOKEN = ""

type MixpanelClient struct {
	token   string
	baseUrl string
}

func TrackCommandUsage(command *cobra.Command) error {
	mixpanelEvent := newEvent(command)
	mixpanelEvent.setUpEvent(MIXPANEL_PROJECT_TOKEN, FLOW_CLI)
	distinctId, err := uniqueUserID()
	if err != nil {
		return err
	}
	mixpanelEvent.setEventDistinctId(distinctId)
	eventPayload, err := encodePayload(mixpanelEvent)
	if err != nil {
		return err
	}
	payload := bytes.NewReader(eventPayload)
	req, err := http.NewRequest("POST", MIXPANEL_TRACK_URL, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "text/plain")
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("invalid response status code %d for tracking command usage", res.StatusCode)
	}

	return nil
}

func encodePayload(obj any) ([]byte, error) {
	b, err := json.Marshal([]any{obj})
	if err != nil {
		return nil, err
	}
	return b, nil
}

func uniqueUserID() (string, error) {
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
