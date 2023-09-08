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

package super

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsSetup struct {
	Scaffold bool `default:"" flag:"scaffold" info:"Use provided scaffolds for project creation"`
}

var setupFlags = flagsSetup{}

var SetupCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "setup <project name>",
		Short:   "Start a new Flow project",
		Example: "flow setup my-project",
		Args:    cobra.ExactArgs(1),
		GroupID: "super",
	},
	Flags: &setupFlags,
	Run:   create,
}

const scaffoldListURL = "https://raw.githubusercontent.com/onflow/flow-cli/master/scaffolds.json"

type scaffold struct {
	Repo        string `json:"repo"`
	Branch      string `json:"branch"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Commit      string `json:"commit"`
	Folder      string `json:"folder"`
	Type        string `json:"type"`
}

func create(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {
	targetDir, err := getTargetDirectory(args[0])
	if err != nil {
		return nil, err
	}

	scaffolds, err := getScaffolds()

	if err != nil {
		return nil, err
	}

	// default to first scaffold - basic scaffold
	pickedScaffold := scaffolds[0]

	if setupFlags.Scaffold {
		scaffoldItems := make([]util.ScaffoldItem, 0)
		for i, s := range scaffolds {
			scaffoldItems = append(
				scaffoldItems,
				util.ScaffoldItem{
					Index:    i,
					Title:    fmt.Sprintf("%s - %s", output.Bold(s.Name), s.Description),
					Category: s.Type,
				},
			)
		}

		selected := util.ScaffoldPrompt(logger, scaffoldItems)
		pickedScaffold = scaffolds[selected]
	}

	logger.StartProgress(fmt.Sprintf("Creating your project %s", targetDir))
	err = cloneScaffold(targetDir, pickedScaffold)
	if err != nil {
		return nil, fmt.Errorf("failed creating scaffold %w", err)
	}
	logger.StopProgress()

	return &setupResult{targetDir: targetDir}, nil
}

func getTargetDirectory(directory string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	target := filepath.Join(pwd, directory)
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
			return "", fmt.Errorf("directory is not empty: %s", target)
		}
	}
	return target, nil
}

func getScaffolds() ([]scaffold, error) {
	httpClient := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest(http.MethodGet, scaffoldListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request for scaffold list: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed requesting scaffold list: %w", err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading scaffold list response: %w", err)
	}

	var all []scaffold
	err = json.Unmarshal(body, &all)
	if err != nil {
		return nil, fmt.Errorf("failed parsing scaffold list response: %w", err)
	}

	valid := make([]scaffold, 0)
	for _, s := range all {
		if s.Repo != "" && s.Description != "" && s.Name != "" && s.Commit != "" {
			valid = append(valid, s)
		}
	}

	return valid, nil
}

func cloneScaffold(targetDir string, conf scaffold) error {
	repo, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL: conf.Repo,
	})
	if err != nil {
		return fmt.Errorf("could not download the scaffold: %w", err)
	}

	worktree, _ := repo.Worktree()
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(conf.Commit),
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("could not find the scaffold version")
	}

	// if we defined a folder remove everything else
	if conf.Folder != "" {
		err = os.Rename(
			filepath.Join(targetDir, conf.Folder),
			filepath.Join(targetDir, "../scaffold-temp"),
		)
		if err != nil {
			return err
		}

		if err = os.RemoveAll(targetDir); err != nil {
			return err
		}

		if err = os.Rename(filepath.Join(targetDir, "../scaffold-temp"), targetDir); err != nil {
			return err
		}
	}

	return os.RemoveAll(filepath.Join(targetDir, ".git"))
}

type setupResult struct {
	targetDir string
}

func (s *setupResult) String() string {
	wd, _ := os.Getwd()
	relDir, _ := filepath.Rel(wd, s.targetDir)
	out := bytes.Buffer{}

	out.WriteString(fmt.Sprintf("%s Congrats! your project was created.\n\n", output.SuccessEmoji()))
	out.WriteString("Start development by following these steps:\n")
	out.WriteString(fmt.Sprintf("1. '%s' to change to your new project,\n", output.Bold(fmt.Sprintf("cd %s", relDir))))
	out.WriteString(fmt.Sprintf("2. '%s' or run Flowser to start the emulator,\n", output.Bold("flow emulator")))
	out.WriteString(fmt.Sprintf("3. '%s' to start developing.\n\n", output.Bold("flow dev")))
	out.WriteString(fmt.Sprintf("You should also read README.md to learn more about the development process!\n"))

	return out.String()
}

func (s *setupResult) Oneliner() string {
	return fmt.Sprintf("Project created inside %s", s.targetDir)
}

func (s *setupResult) JSON() any {
	return nil
}
