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

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		ok    bool
	}{
		{"a@b.co", true},
		{"user@example.com", true},
		{" user@example.com ", true},
		{"", false},
		{"no-at", false},
		{"a@b", false},
		{"@example.com", false},
		{"user@.com", false},
	}
	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if tt.ok && err != nil {
			t.Fatalf("expected ok for %q, got %v", tt.email, err)
		}
		if !tt.ok && err == nil {
			t.Fatalf("expected error for %q", tt.email)
		}
	}
}

