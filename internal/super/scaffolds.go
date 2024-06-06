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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/onflow/flowkit/v2/output"
)

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

func handleScaffold(
	projectName string,
	logger output.Logger,
) (string, error) {
	targetDir, err := getTargetDirectory(projectName)
	if err != nil {
		return "", err
	}

	selectedScaffold, err := selectScaffold(logger)
	if err != nil {
		return "", fmt.Errorf("error selecting scaffold %w", err)
	}

	logger.StartProgress(fmt.Sprintf("Creating your project %s", targetDir))
	defer logger.StopProgress()

	if selectedScaffold != nil {
		err = cloneScaffold(targetDir, *selectedScaffold)
		if err != nil {
			return "", fmt.Errorf("failed creating scaffold %w", err)
		}
	}

	return targetDir, nil
}

func selectScaffold(logger output.Logger) (*scaffold, error) {
	scaffolds, err := getScaffolds()
	if err != nil {
		return nil, err
	}

	// default to first scaffold - basic scaffold
	pickedScaffold := scaffolds[0]

	if setupFlags.ScaffoldID != 0 {
		if setupFlags.ScaffoldID > len(scaffolds) {
			return nil, fmt.Errorf("scaffold with id %d does not exist", setupFlags.ScaffoldID)
		}
		pickedScaffold = scaffolds[setupFlags.ScaffoldID-1]
	}

	if setupFlags.Scaffold {
		scaffoldItems := make([]prompt.ScaffoldItem, 0)
		for i, s := range scaffolds {
			scaffoldItems = append(
				scaffoldItems,
				prompt.ScaffoldItem{
					Index:    i,
					Title:    fmt.Sprintf("%s - %s", output.Bold(s.Name), s.Description),
					Category: s.Type,
				},
			)
		}

		selected := prompt.ScaffoldPrompt(logger, scaffoldItems)
		pickedScaffold = scaffolds[selected]
	}

	return &pickedScaffold, nil
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
