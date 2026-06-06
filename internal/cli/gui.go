// Copyright 2026 Grobmeier Solutions GmbH. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the HumbleBee GUI",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchGUI()
	},
}

type guiLaunchCandidate struct {
	command   string
	args      []string
	checkPath string
}

func launchGUI() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	for _, candidate := range guiLaunchCandidates(executable, os.Getenv("HUMBLEBEE_GUI_PATH"), runtime.GOOS) {
		if candidate.checkPath != "" && !fileExists(candidate.checkPath) {
			continue
		}
		if err := exec.Command(candidate.command, candidate.args...).Start(); err == nil {
			return nil
		}
	}

	return errors.New("HumbleBee GUI was not found. Install the UI release, put humblebee-gui on PATH, or set HUMBLEBEE_GUI_PATH.")
}

func guiLaunchCandidates(cliExecutable string, configuredPath string, goos string) []guiLaunchCandidate {
	var candidates []guiLaunchCandidate
	addPath := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		if goos == "darwin" && strings.HasSuffix(path, ".app") {
			candidates = append(candidates, guiLaunchCandidate{command: "open", args: []string{path}, checkPath: path})
			return
		}
		candidate := guiLaunchCandidate{command: path}
		if strings.ContainsAny(path, `/\`) {
			candidate.checkPath = path
		}
		candidates = append(candidates, candidate)
	}

	addPath(configuredPath)
	dir := filepath.Dir(cliExecutable)
	switch goos {
	case "darwin":
		addPath(filepath.Join(dir, "HumbleBee.app"))
		addPath("/Applications/HumbleBee.app")
		addPath(filepath.Join(dir, "humblebee-gui"))
		addPath("humblebee-gui")
	case "windows":
		addPath(filepath.Join(dir, "HumbleBee.exe"))
		addPath(filepath.Join(dir, "humblebee-gui.exe"))
		addPath("humblebee-gui.exe")
		addPath("HumbleBee.exe")
	default:
		addPath(filepath.Join(dir, "humblebee-gui"))
		addPath(filepath.Join(dir, "HumbleBee"))
		addPath(filepath.Join(dir, "HumbleBee.AppImage"))
		addPath("humblebee-gui")
		addPath("HumbleBee")
	}
	return candidates
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
