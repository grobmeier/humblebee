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

