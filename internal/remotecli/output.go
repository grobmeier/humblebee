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

package remotecli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/grobmeier/humblebee/internal/connected"
	"github.com/spf13/cobra"
)

func safeText(text string) string {
	var output strings.Builder
	for _, r := range text {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			quoted := strconv.QuoteRuneToASCII(r)
			output.WriteString(quoted[1 : len(quoted)-1])
		} else {
			output.WriteRune(r)
		}
	}
	return output.String()
}

func printResult(cmd *cobra.Command, asJSON bool, profile connected.Profile, data any) error {
	if asJSON {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"version": 1, "ok": true, "profile": profile.Name,
			"accountIdentity": profile.AccountIdentity, "timezone": profile.Timezone, "data": data})
	}
	var text strings.Builder
	fmt.Fprintf(&text, "Profile: %s | Account: %s | Timezone: %s\n", safeText(profile.Name), safeText(profile.AccountIdentity), safeText(profile.Timezone))
	switch v := data.(type) {
	case connected.Capabilities:
		fmt.Fprintf(&text, "Connected contract: %d\nScopes: %s\n", v.ContractVersion, safeText(strings.Join(v.Scopes, ", ")))
	case []connected.Project:
		for _, p := range v {
			fmt.Fprintf(&text, "%s  %s  active=%t\n", p.UUID, safeText(p.Name), p.Active)
		}
	case connected.ProjectTasks:
		fmt.Fprintf(&text, "Project: %s\n", safeText(v.ProjectName))
		for _, t := range v.Tasks {
			fmt.Fprintf(&text, "%s  %s  active=%t completed=%t\n", t.UUID, safeText(t.Name), t.Active, t.Completed)
		}
	case connected.Day:
		zone, _ := time.LoadLocation(profile.Timezone)
		fmt.Fprintf(&text, "Date: %s\n", v.Date)
		for _, e := range v.Entries {
			fmt.Fprintf(&text, "%s  %s - %s  %s  %s / %s\n  %s\n", e.UUID, time.Unix(e.StartAtTS, 0).In(zone).Format("15:04"), time.Unix(e.EndAtTS, 0).In(zone).Format("15:04"), formatDuration(e.DurationMS), safeText(e.ProjectName), safeText(e.TaskName), safeText(e.Note))
		}
		fmt.Fprintf(&text, "Total: %s\n", formatDuration(v.TotalDurationMS))
	case connected.StopwatchState:
		zone, _ := time.LoadLocation(profile.Timezone)
		for _, watch := range v.Stopwatches {
			state := "stopped"
			if watch.Running() {
				state = "running"
			}
			start := "not started"
			if watch.StartTime != nil {
				start = time.Unix(*watch.StartTime, 0).In(zone).Format("2006-01-02 15:04:05")
			}
			fmt.Fprintf(&text, "%d  %s / %s  %s  %s  %s\n", watch.ID, safeText(watch.ProjectName), safeText(watch.TaskName), state, start, formatDuration(watch.DurationMS))
		}
	case map[string]string:
		fmt.Fprintln(&text, safeText(v["message"]))
	case connected.WriteResult:
		fmt.Fprintf(&text, "Applied %s | Mutation: %s\n", safeText(v.Mutation.Operation), v.Mutation.ID)
		if v.Entry != nil {
			fmt.Fprintf(&text, "Time entry: %s | Duration: %s\n", v.Entry.UUID, formatDuration(v.Entry.DurationMS))
		}
		if v.Stopwatch != nil {
			fmt.Fprintf(&text, "Stopwatch: %d | Duration: %s\n", v.Stopwatch.ID, formatDuration(v.Stopwatch.DurationMS))
		}
		if v.StopwatchID > 0 {
			fmt.Fprintf(&text, "Discarded stopwatch: %d (no time booked)\n", v.StopwatchID)
		}
	case connected.Mutation:
		fmt.Fprintf(&text, "Mutation: %s | State: %s | Operation: %s\n", v.ID, safeText(v.State), safeText(v.Operation))
		if v.Resources.TimeUUID != "" {
			fmt.Fprintf(&text, "Time entry: %s\n", v.Resources.TimeUUID)
		}
		if v.Resources.StopwatchID > 0 {
			fmt.Fprintf(&text, "Stopwatch: %d\n", v.Resources.StopwatchID)
		}
	default:
		return connected.Failure("client.output", "Unsupported remote output.", 5)
	}
	_, err := io.WriteString(cmd.OutOrStdout(), text.String())
	return err
}

func formatDuration(ms int64) string {
	seconds := ms / 1000
	return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds/60%60, seconds%60)
}

func IsRemote(cmd *cobra.Command) bool {
	for cmd != nil {
		if cmd.Name() == "remote" {
			return true
		}
		cmd = cmd.Parent()
	}
	return false
}

// ReportError is called once at the executable boundary, including parser errors.
func ReportError(cmd *cobra.Command, err error) error {
	var problem *connected.Error
	if !errors.As(err, &problem) {
		problem = connected.Failure("client.invalid_arguments", "Remote command failed. Check its arguments and --help.", 2)
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	if asJSON {
		_ = json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"version": 1, "ok": false, "error": problem})
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s (%s)\n", safeText(problem.Message), problem.Code)
		if problem.MutationID != "" {
			profile, _ := cmd.Flags().GetString("profile")
			profileFlag := ""
			if profile != "" {
				profileFlag = " --profile " + safeText(profile)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Mutation: %s\nRecover: humblebee remote operation%s %s\n", problem.MutationID, profileFlag, problem.MutationID)
		}
	}
	return problem
}
