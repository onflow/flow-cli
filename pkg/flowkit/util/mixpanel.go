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
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	MIXPANEL_TRACK_URL = "https://api.mixpanel.com/track"
)

type MixpanelClient struct {
	token   string
	baseUrl string
}

func SendEvent(command string) error {
	mixpanelev := NewEvent(command)
	mixpanelev.SetUpEvent(MIXPANEL_PROJECT_TOKEN, FLOW_CLI)
	eventPayload, _ := encodePayload(mixpanelev)
	payload := strings.NewReader(eventPayload)
	req, _ := http.NewRequest("POST", MIXPANEL_TRACK_URL, payload)
	req.Header.Add("Accept", "text/plain")
	req.Header.Add("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Printf("mixpanel result")
	fmt.Println(res)
	fmt.Println(string(body))
	return nil

}
func encodePayload(obj interface{}) (string, error) {
	//b, err := json.MarshalIndent(obj, "", "  ")
	//if err != nil {
	//	return "", err
	//}
	//return base64.StdEncoding.EncodeToString(b), nil
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	//return base64.StdEncoding.EncodeToString(b), nil
	formattedString := "[" + string(b) + "]"
	fmt.Printf("mixpanel event payload:\n%v\n", formattedString)
	return formattedString, nil
}
