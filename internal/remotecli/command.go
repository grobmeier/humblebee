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
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grobmeier/humblebee/internal/connected"
	"github.com/grobmeier/humblebee/internal/paths"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type Options struct {
	Store *connected.Store
	Token func() string
	HTTP  connected.Options
	Now   func() time.Time
}

type commandState struct {
	options Options
	profile string
	json    bool
}

func NewCommand(options Options) *cobra.Command {
	if options.Token == nil {
		options.Token = func() string { return os.Getenv("HUMBLEBEE_TAB_TOKEN") }
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	state := &commandState{options: options}
	root := &cobra.Command{Use: "remote", Short: "Use an online Time & Bill account without changing local data", SilenceErrors: true, SilenceUsage: true}
	root.Args = noArgs
	root.RunE = func(cmd *cobra.Command, _ []string) error { return cmd.Help() }
	root.PersistentFlags().StringVar(&state.profile, "profile", "", "Named remote profile")
	root.PersistentFlags().BoolVar(&state.json, "json", false, "Print one versioned JSON result")
	root.SetFlagErrorFunc(func(*cobra.Command, error) error { return invalidArguments() })
	root.AddCommand(state.connectCommand(), state.disconnectCommand())
	root.AddCommand(state.readCommand("status", func(ctx context.Context, s *connected.Session, _ *cobra.Command) (any, error) {
		return s.Capabilities, nil
	}))
	var allProjects bool
	projects := state.readCommand("projects", func(ctx context.Context, s *connected.Session, _ *cobra.Command) (any, error) {
		return s.Client.Projects(ctx, allProjects)
	})
	projects.Flags().BoolVar(&allProjects, "all", false, "Include inactive and special projects")
	root.AddCommand(projects)
	var project string
	var allTasks bool
	tasks := state.readCommand("tasks", func(ctx context.Context, s *connected.Session, _ *cobra.Command) (any, error) {
		return s.Client.Tasks(ctx, project, allTasks)
	})
	tasks.Flags().StringVar(&project, "project", "", "Project UUID")
	tasks.Flags().BoolVar(&allTasks, "all", false, "Include completed and inactive tasks")
	_ = tasks.MarkFlagRequired("project")
	root.AddCommand(tasks)
	var date string
	day := state.readCommand("day", func(ctx context.Context, s *connected.Session, _ *cobra.Command) (any, error) {
		if date == "" {
			zone, _ := time.LoadLocation(s.Capabilities.Timezone)
			date = state.options.Now().In(zone).Format("2006-01-02")
		}
		return s.Client.Day(ctx, date)
	})
	day.Flags().StringVar(&date, "date", "", "Account-local date (YYYY-MM-DD); default today")
	root.AddCommand(day)
	root.AddCommand(state.readCommand("stopwatches", func(ctx context.Context, s *connected.Session, _ *cobra.Command) (any, error) {
		return s.Client.Stopwatches(ctx)
	}))
	root.AddCommand(state.bookCommand(), state.startCommand(), state.stopCommand(), state.discardCommand(), state.operationCommand())
	return root
}

func invalidArguments() error {
	return connected.Failure("client.invalid_arguments", "Invalid remote arguments. See humblebee remote --help.", 2)
}

func (s *commandState) store() (*connected.Store, error) {
	if s.options.Store != nil {
		return s.options.Store, nil
	}
	dir, err := paths.DataDir()
	if err != nil {
		return nil, connected.Failure("client.profile_directory", "Cannot locate remote configuration directory.", 5)
	}
	return connected.NewStore(filepath.Join(dir, "remote"), connected.NativeCredentials{}), nil
}

func (s *commandState) open(cmd *cobra.Command) (*connected.Session, error) {
	store, err := s.store()
	if err != nil {
		return nil, err
	}
	token := s.options.Token()
	if token != "" && s.profile == "" {
		return nil, connected.Failure("client.profile_required", "Use --profile explicitly with HUMBLEBEE_TAB_TOKEN.", 2)
	}
	return store.Open(cmd.Context(), s.profile, token, s.options.HTTP)
}

func noArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return invalidArguments()
	}
	return nil
}

func (s *commandState) readCommand(name string, read func(context.Context, *connected.Session, *cobra.Command) (any, error)) *cobra.Command {
	return &cobra.Command{Use: name, Short: "Read remote " + name, Args: noArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		session, err := s.open(cmd)
		if err != nil {
			return err
		}
		data, err := read(cmd.Context(), session, cmd)
		if err != nil {
			return err
		}
		return printResult(cmd, s.json, session.Profile, data)
	}}
}

func (s *commandState) connectCommand() *cobra.Command {
	var name, origin string
	var ephemeral, replace, stdin, loopback bool
	cmd := &cobra.Command{Use: "connect", Short: "Verify and save a named Time & Bill connection", Args: noArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		store, err := s.store()
		if err != nil {
			return err
		}
		token := s.options.Token()
		if token != "" && (!ephemeral || stdin) {
			return connected.Failure("client.token_source", "Use --ephemeral with HUMBLEBEE_TAB_TOKEN and select only one token source.", 2)
		}
		if stdin {
			data, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), 4097))
			if err != nil || len(data) > 4096 {
				return connected.Failure("client.token_input", "Cannot read a bounded app token from stdin.", 2)
			}
			token = strings.TrimRight(string(data), "\r\n")
		}
		if token == "" {
			if ephemeral || !term.IsTerminal(int(os.Stdin.Fd())) {
				return connected.Failure("client.token_required", "Provide --token-stdin, or HUMBLEBEE_TAB_TOKEN with --ephemeral. Interactive prompts require a terminal.", 3)
			}
			_, _ = io.WriteString(cmd.ErrOrStderr(), "Time & Bill app token: ")
			data, err := term.ReadPassword(int(os.Stdin.Fd()))
			_, _ = io.WriteString(cmd.ErrOrStderr(), "\n")
			if err != nil || len(data) > 4096 {
				return connected.Failure("client.token_input", "Cannot read the app token securely.", 3)
			}
			token = string(data)
		}
		options := s.options.HTTP
		options.AllowLoopbackHTTP = loopback
		p, err := store.Connect(cmd.Context(), name, origin, token, ephemeral, replace, options)
		if err != nil {
			return err
		}
		return printResult(cmd, s.json, p, map[string]string{"message": "Connected. Local databases and commands are unchanged."})
	}}
	cmd.Flags().StringVar(&name, "name", "", "New profile name")
	cmd.Flags().StringVar(&origin, "url", "", "HTTPS Time & Bill server origin")
	cmd.Flags().BoolVar(&ephemeral, "ephemeral", false, "Do not store the token; provide it again for subsequent commands")
	cmd.Flags().BoolVar(&replace, "replace", false, "Confirm replacement of an existing profile's server/account binding")
	cmd.Flags().BoolVar(&stdin, "token-stdin", false, "Read the app token from stdin, never from an argument")
	cmd.Flags().BoolVar(&loopback, "allow-loopback-http", false, "Allow unencrypted HTTP only for a loopback development server")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

func (s *commandState) disconnectCommand() *cobra.Command {
	return &cobra.Command{Use: "disconnect", Short: "Remove a local remote profile without changing server data", Args: noArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		store, err := s.store()
		if err != nil {
			return err
		}
		p, err := store.Select(s.profile)
		if err != nil {
			return err
		}
		if err := store.Disconnect(p.Name); err != nil {
			return err
		}
		return printResult(cmd, s.json, p, map[string]string{"message": "Local profile and stored credential removed. Server timers and data are unchanged. Revoke the app token separately in Time & Bill."})
	}}
}
