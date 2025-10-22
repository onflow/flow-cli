/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

package emulator

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"sync"

	"github.com/dukex/mixpanel"
	"github.com/onflow/flow-emulator/cmd/emulator/start"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"

	"github.com/onflow/flow-cli/build"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/settings"
	"github.com/onflow/flow-cli/internal/util"
)

var Cmd *cobra.Command

// Mixpanel client to be reused on each http request of the middleware
var mixpanelClient mixpanel.Mixpanel

func configuredServiceKey(
	init bool,
	_ crypto.SignatureAlgorithm,
	_ crypto.HashAlgorithm,
) (
	crypto.PrivateKey,
	crypto.SignatureAlgorithm,
	crypto.HashAlgorithm,
) {
	var state *flowkit.State
	var err error
	loader := &afero.Afero{Fs: afero.NewOsFs()}
	command.UsageMetrics(Cmd, &sync.WaitGroup{})

	if init {
		state, err = flowkit.Init(loader)
		if err != nil {
			exitf(1, err.Error())
		} else {
			err = state.SaveDefault()
			if err != nil {
				exitf(1, err.Error())
			}
		}
	} else {
		state, err = flowkit.Load(command.Flags.ConfigPaths, loader)
		if err != nil {
			if errors.Is(err, config.ErrDoesNotExist) {
				exitf(1, "🙏 Configuration (flow.json) is missing, are you in the correct directory? If you are trying to create a new project, initialize it with 'flow init' and then rerun this command.")
			} else {
				exitf(1, err.Error())
			}
		}
	}

	serviceAccount, err := state.EmulatorServiceAccount()
	if err != nil {
		util.Exit(1, err.Error())
	}

	privateKey, err := serviceAccount.Key.PrivateKey()
	if err != nil {
		util.Exit(1, "Only hexadecimal keys can be used as the emulator service account key.")
	}

	err = serviceAccount.Key.Validate()
	if err != nil {
		util.Exit(
			1,
			fmt.Sprintf("invalid private key in %s emulator configuration, %s",
				serviceAccount.Name,
				err.Error(),
			),
		)
	}

	return *privateKey, serviceAccount.Key.SigAlgo(), serviceAccount.Key.HashAlgo()
}

func trackRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a unique user ID
		usr, _ := user.Current() // ignore err, just use empty string
		hash := sha256.Sum256(fmt.Appendf(nil, "%s%s", usr.Username, usr.Uid))
		userID := base64.StdEncoding.EncodeToString(hash[:])

		// Track the request in Mixpanel
		_ = mixpanelClient.Track(userID, "emulator-request", &mixpanel.Event{
			IP: "0", // do not track IPs
			Properties: map[string]any{
				"method":  r.Method,
				"url":     r.URL.String(),
				"version": build.Semver(),
				"os":      runtime.GOOS,
				"ci":      os.Getenv("CI") != "", // CI is commonly set by CI providers
			},
		})

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func init() {
	// Initialize mixpanel client only if metrics are enabled and token is not empty
	if settings.MetricsEnabled() && command.MixpanelToken != "" {
		mixpanelClient = mixpanel.New(command.MixpanelToken, "")
		Cmd = start.Cmd(start.StartConfig{
			GetServiceKey:   configuredServiceKey,
			RestMiddlewares: []start.HttpMiddleware{trackRequestMiddleware},
		})
	} else {
		Cmd = start.Cmd(start.StartConfig{
			GetServiceKey:   configuredServiceKey,
			RestMiddlewares: []start.HttpMiddleware{},
		})
	}

	// Add --fork flag with optional value (defaults to mainnet when value omitted)
	Cmd.Flags().String("fork", "", "fork from a remote network defined in flow.json (typically mainnet or testnet). If provided without a value, defaults to mainnet")
	if f := Cmd.Flags().Lookup("fork"); f != nil {
		f.NoOptDefVal = "mainnet"
	}

	Cmd.Use = "emulator"
	Cmd.Short = "Run Flow network for development"
	Cmd.GroupID = "tools"
	SnapshotCmd.AddToParent(Cmd)

	// Translate --fork to --fork-host before emulator reads flags
	Cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		fh, err := cmd.Flags().GetString("fork-host")
		if err != nil {
			return err
		}
		if fh != "" {
			return nil
		}
		forkOpt, err := cmd.Flags().GetString("fork")
		if err != nil {
			return err
		}
		if forkOpt == "" {
			return nil
		}
		loader := &afero.Afero{Fs: afero.NewOsFs()}
		state, err := flowkit.Load(command.Flags.ConfigPaths, loader)
		if err != nil {
			return fmt.Errorf("failed to load flow.json: %w", err)
		}

		// Resolve network endpoint from flow.json
		network, err := state.Networks().ByName(forkOpt)
		if err != nil {
			return fmt.Errorf("network %q not found in flow.json", forkOpt)
		}
		host := network.Host
		if host == "" {
			return fmt.Errorf("network %q has no host configured", forkOpt)
		}

		// Detect chain ID from the network
		chainID, err := util.GetNetworkChainID(state, forkOpt)
		if err != nil {
			return err
		}

		// Set fork-host flag
		if err := cmd.Flags().Set("fork-host", host); err != nil {
			return err
		}

		// Automatically disable signature validation when forking
		// This is necessary because forked transactions were signed for the original network
		if err := cmd.Flags().Set("skip-tx-validation", "true"); err != nil {
			return err
		}
		
		// Log info to stderr
		fmt.Fprintf(os.Stderr, "Forking from %s (chain ID: %s)\n", host, chainID)
		fmt.Fprintf(os.Stderr, "Signature validation automatically disabled for fork mode\n")

		return nil
	}
}

func exitf(code int, msg string, args ...any) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(code)
}
