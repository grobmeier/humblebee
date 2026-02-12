package validator

import (
	"errors"
	"strings"
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > 254 {
		return errors.New("email is invalid")
	}
	at := strings.Index(email, "@")
	if at <= 0 || at >= len(email)-3 {
		return errors.New("email is invalid")
	}
	dot := strings.LastIndex(email, ".")
	if dot < at+2 || dot >= len(email)-1 {
		return errors.New("email is invalid")
	}
	return nil
}

