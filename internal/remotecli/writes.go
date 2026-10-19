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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/grobmeier/humblebee/internal/connected"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func (s *commandState) writeCommand(name string, write func(context.Context, *connected.Session, *connected.Store, *cobra.Command) (any, error)) *cobra.Command {
	return &cobra.Command{Use: name, Short: "Perform an explicit remote " + name + " operation", Args: noArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		session, err := s.open(cmd)
		if err != nil {
			return err
		}
		store, err := s.store()
		if err != nil {
			return err
		}
		result, err := write(cmd.Context(), session, store, cmd)
		if err != nil {
			return err
		}
		return printResult(cmd, s.json, session.Profile, result, session.Client.RedactText)
	}}
}

func (s *commandState) bookCommand() *cobra.Command {
	var b connected.Booking
	var noteFile string
	cmd := s.writeCommand("book", func(ctx context.Context, session *connected.Session, store *connected.Store, cmd *cobra.Command) (any, error) {
		if noteFile != "" {
			reader := cmd.InOrStdin()
			if noteFile != "-" {
				file, err := os.Open(noteFile)
				if err != nil {
					return nil, connected.Failure("client.note_file", "Cannot open the note file.", 2)
				}
				defer file.Close()
				reader = file
			}
			data, err := io.ReadAll(io.LimitReader(reader, 4097))
			if err != nil || len(data) > 4096 {
				return nil, connected.Failure("client.note_file", "The UTF-8 note file is unreadable or too large.", 2)
			}
			b.Note = string(data)
		}
		return session.Book(ctx, store, b)
	})
	cmd.Flags().StringVar(&b.ProjectUUID, "project", "", "Project UUID")
	cmd.Flags().StringVar(&b.TaskUUID, "task", "", "Task UUID")
	cmd.Flags().StringVar(&b.Date, "date", "", "Account-local booking date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&b.Start, "start", "", "Start time (24-hour HH:mm)")
	cmd.Flags().StringVar(&b.End, "end", "", "End time (24-hour HH:mm; 24:00 is midnight)")
	cmd.Flags().StringVar(&b.Note, "note", "", "Optional note")
	cmd.Flags().StringVar(&noteFile, "note-file", "", "Read a UTF-8 note from a file, or - for stdin")
	for _, name := range []string{"project", "task", "date", "start", "end"} {
		_ = cmd.MarkFlagRequired(name)
	}
	cmd.MarkFlagsMutuallyExclusive("note", "note-file")
	return cmd
}

func (s *commandState) startCommand() *cobra.Command {
	var task string
	cmd := s.writeCommand("start", func(ctx context.Context, session *connected.Session, store *connected.Store, _ *cobra.Command) (any, error) {
		return session.Start(ctx, store, task)
	})
	cmd.Flags().StringVar(&task, "task", "", "Task UUID to start; never switches another running task")
	_ = cmd.MarkFlagRequired("task")
	return cmd
}

func (s *commandState) stopCommand() *cobra.Command {
	var id int64
	cmd := s.writeCommand("stop", func(ctx context.Context, session *connected.Session, store *connected.Store, _ *cobra.Command) (any, error) {
		return session.Stop(ctx, store, id)
	})
	cmd.Flags().Int64Var(&id, "id", 0, "Stopwatch ID; may be omitted only for exactly one running timer")
	return cmd
}

func confirmDiscard(cmd *cobra.Command, id int64, yes bool) error {
	if yes {
		return nil
	}
	in, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(in.Fd())) {
		return connected.Failure("client.confirmation_required", "Discard requires --yes in a non-interactive session. No time was discarded.", 2)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Discard remote stopwatch %d without booking its time? [y/N]: ", id)
	answer, err := bufio.NewReader(io.LimitReader(in, 32)).ReadString('\n')
	if err != nil || strings.ToLower(strings.TrimSpace(answer)) != "y" {
		return connected.Failure("client.cancelled", "Discard cancelled; no time was discarded.", 2)
	}
	return nil
}

func (s *commandState) discardCommand() *cobra.Command {
	var id int64
	var yes bool
	cmd := s.writeCommand("discard", func(ctx context.Context, session *connected.Session, store *connected.Store, cmd *cobra.Command) (any, error) {
		if id <= 0 {
			return nil, connected.Failure("client.invalid_timer", "Choose the exact positive stopwatch --id.", 2)
		}
		if err := confirmDiscard(cmd, id, yes); err != nil {
			return nil, err
		}
		return session.Discard(ctx, store, id)
	})
	cmd.Flags().Int64Var(&id, "id", 0, "Exact stopwatch ID to discard without booking")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm discarding this stopwatch without booking")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (s *commandState) operationCommand() *cobra.Command {
	cmd := s.writeCommand("operation", func(ctx context.Context, session *connected.Session, store *connected.Store, cmd *cobra.Command) (any, error) {
		return session.Operation(ctx, store, cmd.Flags().Arg(0))
	})
	cmd.Use = "operation <mutation-id>"
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return invalidArguments()
		}
		return nil
	}
	return cmd
}
