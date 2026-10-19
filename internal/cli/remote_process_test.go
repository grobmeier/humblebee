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
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRemoteProcessHelper(t *testing.T) {
	if os.Getenv("HUMBLEBEE_CLI_TEST_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"humblebee"}, os.Args[i+1:]...)
			break
		}
	}
	os.Exit(ExitCode(Execute()))
}

func TestRemoteProcessErrorsRemainJSONAndNeverInitializeLocalDatabase(t *testing.T) {
	for _, args := range [][]string{
		{"remote", "status", "--profile", "missing", "--json"},
		{"remote", "status", "--bad", "--json"},
		{"remote", "not-a-command", "--json"},
	} {
		t.Run(args[1]+args[2], func(t *testing.T) {
			home := t.TempDir()
			cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestRemoteProcessHelper$", "--"}, args...)...)
			cmd.Env = append(os.Environ(), "HUMBLEBEE_CLI_TEST_HELPER=1", "HUMBLEBEE_HOME="+home, "HUMBLEBEE_TAB_TOKEN=")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			output, err := cmd.Output()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 2 {
				t.Fatalf("unexpected exit %v: %s %s", err, output, stderr.String())
			}
			var result struct {
				Version int  `json:"version"`
				OK      bool `json:"ok"`
			}
			if json.Unmarshal(output, &result) != nil || result.Version != 1 || result.OK || stderr.Len() != 0 {
				t.Fatalf("invalid error output: %s %s", output, stderr.String())
			}
			if _, err := os.Stat(filepath.Join(home, "humblebee.db")); !os.IsNotExist(err) {
				t.Fatal("remote error created local database")
			}
		})
	}
}
