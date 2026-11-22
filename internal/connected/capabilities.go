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

import (
	"context"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode"
)

type Capabilities struct {
	ContractVersion        int      `json:"contractVersion"`
	AccountIdentity        string   `json:"accountIdentity"`
	Timezone               string   `json:"timezone"`
	Scopes                 []string `json:"scopes"`
	TimestampUnit          string   `json:"timestampUnit"`
	DurationUnit           string   `json:"durationUnit"`
	MutationRecovery       bool     `json:"mutationRecovery"`
	RecoveryWindowSeconds  int64    `json:"recoveryWindowSeconds"`
	ConditionalStopwatches bool     `json:"conditionalStopwatches"`
}

func (c *Client) Capabilities(ctx context.Context) (Capabilities, error) {
	var caps Capabilities
	if err := c.call(ctx, "get_client_capabilities", map[string]any{}, &caps); err != nil {
		return caps, err
	}
	_, zoneErr := time.LoadLocation(caps.Timezone)
	unsafeIdentity := len(caps.AccountIdentity) > 256 || strings.Contains(caps.AccountIdentity, c.token)
	for _, r := range caps.AccountIdentity {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			unsafeIdentity = true
		}
	}
	for _, scope := range caps.Scopes {
		if !safeCode(strings.ReplaceAll(scope, ":", ".")) || strings.Contains(scope, c.token) {
			unsafeIdentity = true
		}
	}
	if caps.ContractVersion != 1 || caps.AccountIdentity == "" || caps.Timezone == "" || zoneErr != nil || caps.Scopes == nil ||
		caps.TimestampUnit != "seconds" || caps.DurationUnit != "milliseconds" || !caps.MutationRecovery ||
		caps.RecoveryWindowSeconds < 7*24*60*60 || !caps.ConditionalStopwatches || unsafeIdentity {
		return Capabilities{}, Failure("client.incompatible", "This server does not provide the required connected-client safety contract.", 7)
	}
	return caps, nil
}
