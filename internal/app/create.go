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

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type FlagsCreate struct{}

var createFlags = FlagsCreate{}

var CreateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "create <path>",
		Short:   "Create a new Flow project",
		Example: "flow app create my-app",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &createFlags,
	Run:   create,
}

const ScaffoldRepo = "https://github.com/onflow/flow-app-scaffold/"

type Examples map[string]json.RawMessage
type ExampleConfig struct {
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

func create(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
) (command.Result, error) {
	var example string
	var template string

	target, err := getTargetDirectory(args[0])
	if err != nil {
		return nil, err
	}

	typePrompt := promptui.Select{
		Label: "Would you like to start from a full featured example or from a template without much code?",
		Items: []string{"Template", "Example"},
	}
	_, useTemplate, err := typePrompt.Run()
	if err != nil {
		return nil, err
	}

	if useTemplate == "Example" {
		example, err = createFromExample(target)
	} else {
		template, err = createFromTemplate(target)
	}
	if err != nil {
		return nil, err
	}

	if example != "" {
		return &CreateResult{
			created: target,
			example: example,
		}, nil
	} else {
		return &CreateResult{
			created:  target,
			template: template,
		}, nil
	}
}

func getTargetDirectory(directory string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	target := path.Join(pwd, directory)
	info, err := os.Stat(target)
	if !os.IsNotExist(err) {
		if !info.IsDir() {
			return "", fmt.Errorf("%s is a file", target)
		}

		file, err := os.Open(target)
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = file.Readdirnames(1)
		if err != io.EOF {
			return "", fmt.Errorf("directory is not empty or can't be accessed: %s", target)
		}
	}
	return target, nil
}

func askChoice(
	target string,
	message string,
) (string, error) {
	folders, err := os.ReadDir(target)
	if err != nil {
		return "", err
	}

	choices := make([]string, 0)
	for _, folder := range folders {
		if folder.IsDir() && !strings.HasPrefix(folder.Name(), ".") {
			choices = append(choices, folder.Name())
		}
	}

	prompt := promptui.Select{
		Label: message,
		Items: choices,
	}
	_, value, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return value, nil
}

func createFromExample(target string) (string, error) {
	url := ScaffoldRepo + "blob/main/examples.json?raw=1"
	httpClient := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	examples := Examples{}
	err = json.Unmarshal(body, &examples)
	if err != nil {
		return "", err
	}

	options := reflect.ValueOf(examples).MapKeys()
	prompt := promptui.Select{
		Label: "Which example do you want to start with?",
		Items: options,
	}
	_, option, err := prompt.Run()
	if err != nil {
		return "", err
	}

	config := ExampleConfig{}
	err = json.Unmarshal(examples[option], &config)

	if err != nil {
		return "", err
	}

	_, err = git.PlainClone(target, false, &git.CloneOptions{
		URL:           config.Repo,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", config.Branch)),
		SingleBranch:  true,
		Progress:      os.Stdout,
	})
	if err != nil {
		return "", err
	}

	return option, nil
}

func createFromTemplate(target string) (string, error) {
	_, err := git.PlainClone(target, false, &git.CloneOptions{
		URL:      ScaffoldRepo,
		Progress: os.Stdout,
	})
	if err != nil {
		return "", err
	}

	template, err := askChoice(target, "Which template do you want to start with?")
	if err != nil {
		return "", err
	}

	err = cleanUpProject(target, template)
	if err != nil {
		return "", err
	}
	return template, nil
}

func cleanUpProject(target string, template string) error {
	entries, err := os.ReadDir(target)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Name() == template {
			continue
		}

		if entry.IsDir() {
			err = os.RemoveAll(path.Join(target, entry.Name()))
		} else {
			err = os.Remove(path.Join(target, entry.Name()))
		}
		if err != nil {
			return err
		}
	}

	files, err := os.ReadDir(path.Join(target, template))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.Rename(path.Join(target, template, file.Name()), path.Join(target, file.Name()))
		if err != nil {
			return err
		}
	}

	err = os.Remove(path.Join(target, template))
	if err != nil {
		return err
	}

	return nil
}
