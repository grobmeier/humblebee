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

import "testing"

func TestGUILaunchCandidatesPreferConfiguredPath(t *testing.T) {
	candidates := guiLaunchCandidates("/usr/local/bin/humblebee", "/Applications/HumbleBee.app", "darwin")

	if len(candidates) == 0 {
		t.Fatal("expected launch candidates")
	}
	if candidates[0].command != "open" || len(candidates[0].args) != 1 || candidates[0].args[0] != "/Applications/HumbleBee.app" {
		t.Fatalf("expected configured macOS app to be opened first, got %#v", candidates[0])
	}
}

func TestGUILaunchCandidatesIncludeAdjacentBinary(t *testing.T) {
	candidates := guiLaunchCandidates("/usr/local/bin/humblebee", "", "linux")

	if len(candidates) == 0 {
		t.Fatal("expected launch candidates")
	}
	if candidates[0].command != "/usr/local/bin/humblebee-gui" {
		t.Fatalf("expected adjacent GUI binary first, got %#v", candidates[0])
	}
}
