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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	MIXPANEL_TRACK_URL              = "https://api.mixpanel.com/track"
	MIXPANEL_QUERY_URL              = "https://mixpanel.com/api/2.0/engage?project_id=2763737"
	MIXPANEL_PROFILE_URL            = "https://api.mixpanel.com/engage#profile-set"
	MIXPANEL_PROJECT_TOKEN          = "7af4e6f44df2c77935477ba103b3c529"
	MIXPANEL_SERVICE_ACCOUNT_SECRET = "Rmxvdy1jbGkuZTVkMTNjLm1wLXNlcnZpY2UtYWNjb3VudDp2TkZrVzhiWWNSY1ZuQmtMZFF4bXVzamdZa0dyc2FsMQ=="
)

type MixpanelClient struct {
	token   string
	baseUrl string
}

func TrackCommandUsage(command *cobra.Command) error {
	mixpanelEvent := newEvent(command)
	mixpanelEvent.setUpEvent(MIXPANEL_PROJECT_TOKEN, FLOW_CLI)
	eventPayload, err := encodePayload(mixpanelEvent)
	if err != nil {
		return err
	}
	payload := strings.NewReader(eventPayload)
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

func encodePayload(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	formattedString := "[" + string(b) + "]"
	return formattedString, nil
}

func SetUserTrackingSettings(enable bool) error {
	mixpanelUser, err := getMixPanelUser()
	if err != nil {
		return err
	}
	mixpanelUser.configureUserTracking(enable)

	userPayload, err := encodePayload(mixpanelUser)
	if err != nil {
		return err
	}
	payload := strings.NewReader(userPayload)
	req, err := http.NewRequest("POST", MIXPANEL_PROFILE_URL, payload)
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

type MixPanelResponse struct {
	Results []struct {
		Properties struct {
			OptIn bool `json:"opt_in"`
		} `json:"$properties"`
	} `json:"results"`
}

func IsUserOptedIn() (bool, error) {
	distinctId, err := generateNewDistinctId()
	if err != nil {
		return false, err
	}

	queryPayload := "distinct_id=" + url.QueryEscape(distinctId)
	payload := strings.NewReader(queryPayload)
	req, err := http.NewRequest("POST", MIXPANEL_QUERY_URL, payload)
	if err != nil {
		return false, err
	}

	mixpanelAuth := "Basic " + MIXPANEL_SERVICE_ACCOUNT_SECRET
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", mixpanelAuth)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var queryResponse MixPanelResponse
	err = json.Unmarshal(body, &queryResponse)

	if err != nil {
		return false, err
	}
	if res.StatusCode >= 400 {
		return false, fmt.Errorf("invalid response status code %d for tracking command usage", res.StatusCode)
	}
	if len(queryResponse.Results) == 0 {
		return false, fmt.Errorf("invalid response from Mixpanel Query API")
	}
	return queryResponse.Results[0].Properties.OptIn, nil
}
