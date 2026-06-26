// Copyright (C) 2025 Nguyen Nhat Tung
// This file is part of EVM RPC Agent
//
// EVM RPC Agent is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// EVM RPC Agent is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with EVM RPC Agent. If not, see <https://www.gnu.org/licenses/>.

package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
)

var majorVersion = 0
var minorVersion = 1
var patchVersion = 0
var gitCommit, gitDate, gitBranch string

func version() string {
	originDate := time.Date(2026, time.June, 7, 0, 0, 0, 0, time.UTC)
	gitDate2, err := time.Parse("20060102", gitDate)
	buildDate := opx.Ternary(err == nil, gitDate2, time.Now().UTC())
	duration := buildDate.Sub(originDate)
	minor := minorVersion
	patch := strconv.Itoa(patchVersion)
	if gitBranch == "master" {
		// do nothing
	} else if gitBranch == "release" {
		minor += 1
		patch = patch + "-rc"
	} else if strings.Contains(gitBranch, "feat/") {
		minor += 1
		patch = patch + "-dev"
	} else {
		patch = strconv.Itoa(patchVersion+1) + "-dev"
	}
	if gitCommit != "" && len(gitCommit) >= 8 {
		return fmt.Sprintf("%d.%d.%s.%d-%s", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000), gitCommit[:8])
	}
	return fmt.Sprintf("%d.%d.%s.%d", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000))
}

// Initialize configurations, loggings for internal modules, and display basic
// information about this invocation.
func InitApp() *Controller {
	ctrl := NewController(true)

	pwd, _ := os.Getwd()
	pwd, _ = filepath.Abs(pwd)
	exec, _ := os.Executable()
	exec, _ = filepath.Abs(exec)

	ctrl.Logger.Info().Msgf("EVM RPC Agent v%s", version())
	gitDate2, _ := time.Parse("20060102", gitDate)
	buildDate := opx.Ternary(gitDate == "", time.Now().UTC(), gitDate2)
	ctrl.Logger.Info().Msgf("Copyright (C) %d Nguyen Nhat Tung", buildDate.Year())
	ctrl.Logger.Info().Msg("Licensed under GPL-3.0 license. See COPYING file along with this program for more details.")
	ctrl.Logger.Info().Msgf("Working directory %s", pwd)
	ctrl.Logger.Info().Msgf("Config directory %s", ctrl.Root.ConfigDir)
	ctrl.Logger.Info().Msgf("Executable file %s", exec)
	ctrl.Logger.Info().Msgf("Portable mode %t", ctrl.Root.IsPortable)
	ctrl.Logger.Info().Msg("-----------------")

	return ctrl
}

// Execute the program.
func Execute() {
	rootCmd := &cobra.Command{
		Use: "evm-rpc-agent",
		Long: fmt.Sprintf(
			"EVM RPC Agent v%s.\n\n"+
				"A lightweight agent for proxying, monitoring, and managing EVM-compatible blockchain RPC requests.",
			version()),
		Short:   "EVM RPC agent for proxying and managing Ethereum JSON-RPC requests",
		Version: version(),
	}

	rootCmd.AddCommand(ConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
