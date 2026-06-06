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

package ui

import (
	"errors"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func PromptString(message string, validate func(string) error) (string, error) {
	var out string
	p := &survey.Input{Message: message}
	opts := []survey.AskOpt{}
	if validate != nil {
		opts = append(opts, survey.WithValidator(func(ans any) error {
			s, ok := ans.(string)
			if !ok {
				return errors.New("invalid input")
			}
			return validate(strings.TrimSpace(s))
		}))
	}
	if err := survey.AskOne(p, &out, opts...); err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func Confirm(message string, defaultNo bool) (bool, error) {
	p := &survey.Confirm{
		Message: message,
		Default: !defaultNo,
	}
	var ok bool
	if err := survey.AskOne(p, &ok); err != nil {
		return false, err
	}
	return ok, nil
}

