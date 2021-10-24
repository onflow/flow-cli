/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
	"fmt"
	"io"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

var CreateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "create <path>",
		Short:   "Create a new Flow project",
		Example: "flow app create my-app",
		Args:    cobra.ExactArgs(1),
	},
	Run: create,
}

type FileAction int

const (
	RemoveAll = iota
	RemoveFiles
	Remove
	Rename
	MoveFiles
)

type Action struct {
	action FileAction
	path   string
	value  string
}

type Template struct {
	api     string
	cadence string
	web     string
}

func create(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
) (command.Result, error) {
	var example string
	var template *Template

	target, err := getTargetDirectory(args[0])
	if err != nil {
		return nil, err
	}

	_, err = git.PlainClone(target, false, &git.CloneOptions{
		URL:      "https://github.com/flyinglimao/flow-app-scaffold",
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}
	actions := []Action{{
		action: Rename,
		path:   "README.project.md",
		value:  "README.md",
	}, {
		action: Remove,
		path:   "CONTRIBUTING.md",
	}, {
		action: RemoveAll,
		path:   ".git",
	}, {
		action: RemoveAll,
		path:   ".github",
	}}

	typePrompt := promptui.Select{
		Label: "Would you like to start from a full featured example or from a template without much code?",
		Items: []string{"Template", "Example"},
	}
	_, useTemplate, err := typePrompt.Run()
	if err != nil {
		return nil, err
	}

	if useTemplate == "Example" {
		actions, example, err = createFromExample(actions, target)
	} else {
		actions, template, err = createFromTemplate(actions, target)
	}
	if err != nil {
		return nil, err
	}
	actions = append(actions, Action{
		action: RemoveAll,
		path:   "example",
	})

	err = executeActions(actions, target)
	if err != nil {
		return nil, err
	}

	if template != nil {
		return &CreateResult{
			created: target,
			api:     template.api,
			cadence: template.cadence,
			web:     template.web,
		}, nil
	} else {
		return &CreateResult{
			created: target,
			example: example,
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
	pathName string,
) (string, error) {
	folders, err := os.ReadDir(path.Join(target, pathName))
	if err != nil {
		return "", err
	}

	choices := []string{}
	for _, folder := range folders {
		choices = append(choices, folder.Name())
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

func executeActions(actions []Action, target string) error {
	for _, action := range actions {
		switch action.action {
		case RemoveAll:
			err := os.RemoveAll(path.Join(target, action.path))
			if err != nil {
				return err
			}
		case RemoveFiles:
			files, err := os.ReadDir(path.Join(target, action.path))
			if err != nil {
				return err
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				err := os.Remove(path.Join(target, action.path, file.Name()))
				if err != nil {
					return err
				}
			}
		case Remove:
			err := os.Remove(path.Join(target, action.path))
			if err != nil {
				return err
			}
		case Rename:
			err := os.Rename(path.Join(target, action.path), path.Join(target, action.value))
			if err != nil {
				return err
			}
		case MoveFiles:
			files, err := os.ReadDir(path.Join(target, action.path))
			if err != nil {
				return err
			}
			for _, file := range files {
				err := os.Rename(path.Join(target, action.path, file.Name()), path.Join(target, action.value, file.Name()))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func createFromExample(actions []Action, target string) ([]Action, string, error) {
	actions = append(actions, Action{
		action: RemoveAll,
		path:   "api",
	}, Action{
		action: RemoveAll,
		path:   "cadence",
	}, Action{
		action: RemoveAll,
		path:   "web",
	}, Action{
		action: RemoveFiles,
		path:   ".",
	})

	example, err := askChoice(target, "Which example you want to start with?", "example")
	if err != nil {
		return nil, "", err
	}

	actions = append(actions, Action{
		action: MoveFiles,
		path:   fmt.Sprintf("example/%s", example),
		value:  ".",
	})
	return actions, example, nil
}

func createFromTemplate(actions []Action, target string) ([]Action, *Template, error) {
	api, err := askChoice(target, "Which API template you want to start with?", "api/templates")
	if err != nil {
		return nil, nil, err
	}
	cadence, err := askChoice(target, "Which Cadence template you want to start with?", "cadence/templates")
	if err != nil {
		return nil, nil, err
	}
	web, err := askChoice(target, "Which Web template you want to start with?", "web/templates")
	if err != nil {
		return nil, nil, err
	}

	actions = append(actions, Action{
		action: MoveFiles,
		path:   fmt.Sprintf("api/templates/%s", api),
		value:  "api",
	}, Action{
		action: RemoveAll,
		path:   "api/templates",
	}, Action{
		action: MoveFiles,
		path:   fmt.Sprintf("cadence/templates/%s", cadence),
		value:  "cadence",
	}, Action{
		action: RemoveAll,
		path:   "cadence/templates",
	}, Action{
		action: MoveFiles,
		path:   fmt.Sprintf("web/templates/%s", web),
		value:  "web",
	}, Action{
		action: RemoveAll,
		path:   "web/templates",
	})

	return actions, &Template{
		api:     api,
		cadence: cadence,
		web:     web,
	}, nil
}
