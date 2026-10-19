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

package connected

import "testing"

func TestServerErrorsUseStableCLIExitCategories(t *testing.T) {
	for code, want := range map[string]int{
		"mcp.scope_required": 3, "billing.subscription_required": 3, "time.forbidden": 3, "stopwatch.forbidden": 3,
		"time.conflict": 4, "time.booking_closed": 4, "stopwatch.stale_state": 4, "projects.task_unavailable": 4,
		"time.invalid_time": 2, "time.ambiguous_local_time": 2, "time.nonexistent_local_time": 2, "time.multi_day_unsupported": 2,
	} {
		if got := ExitCode(remoteError(code)); got != want {
			t.Errorf("%s: got %d, want %d", code, got, want)
		}
	}
}
