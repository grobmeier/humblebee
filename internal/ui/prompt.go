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

