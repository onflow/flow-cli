package super

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

type FlagsSetup struct {
	Scaffold bool `default:"" flag:"scaffold" info:"Use provided scaffolds for project creation"`
}

var setupFlags = FlagsSetup{}

var SetupCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "setup <path>",
		Short:   "Setup a new Flow project",
		Example: "flow setup my-project",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &setupFlags,
	Run:   create,
}

const scaffoldListURL = "https://raw.githubusercontent.com/onflow/flow-scaffold-list/main/scaffold-list.json"

type scaffoldConf struct {
	Repo        string `json:"repo"`
	Branch      string `json:"branch"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func create(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
) (command.Result, error) {
	logger := output.NewStdoutLogger(output.InfoLog)

	targetDir, err := getTargetDirectory(args[0])
	if err != nil {
		return nil, err
	}

	scaffolds, err := getScaffolds()
	if err != nil {
		return nil, err
	}

	// default to first scaffold - basic scaffold
	scaffold := scaffolds[0]

	if setupFlags.Scaffold {
		scaffoldList := make([]string, len(scaffolds))
		for i, s := range scaffolds {
			scaffoldList[i] = fmt.Sprintf("%s - %s", output.Bold(s.Name), s.Description)
		}

		selected := output.ScaffoldPrompt(scaffoldList)
		scaffold = scaffolds[selected]
	}

	logger.StartProgress(fmt.Sprintf("Creating your project %s", targetDir))
	err = cloneScaffold(targetDir, scaffold)
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
			return "", fmt.Errorf("directory is not empty: %s", target)
		}
	}
	return target, nil
}

func getScaffolds() ([]scaffoldConf, error) {
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

	var confs []scaffoldConf
	err = json.Unmarshal(body, &confs)
	if err != nil {
		return nil, fmt.Errorf("failed parsing scaffold list response: %w", err)
	}

	return confs, nil
}

func cloneScaffold(targetDir string, conf scaffoldConf) error {
	_, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL: conf.Repo,
	})

	return err
}

type setupResult struct {
	targetDir string
}

func (s *setupResult) String() string {
	wd, _ := os.Getwd()
	relDir, _ := filepath.Rel(wd, s.targetDir)
	out := bytes.Buffer{}

	out.WriteString(fmt.Sprintf("\n%s Your project was created.\n", output.SuccessEmoji()))
	out.WriteString(fmt.Sprintf(
		"\nUse `%s` to view your project files.\nOpen %s/README.md to learn how to get started!\n",
		output.Bold(fmt.Sprintf("cd %s", relDir)),
		relDir,
	))

	return out.String()
}

func (s *setupResult) Oneliner() string {
	return fmt.Sprintf("Project created inside %s", s.targetDir)
}

func (s *setupResult) JSON() interface{} {
	return nil
}
