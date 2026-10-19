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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

const maxResponseBytes = 2 << 20

type Options struct {
	AllowLoopbackHTTP bool
	HTTPClient        *http.Client
}

type Client struct {
	origin string
	token  string
	http   *http.Client
	nextID atomic.Int64
}

func NormalizeOrigin(raw string, allowLoopback bool) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" || (u.Path != "" && u.Path != "/") {
		return "", Failure("client.invalid_origin", "Use a server origin without credentials, a path, query or fragment.", 2)
	}
	host := strings.ToLower(u.Hostname())
	loopback := host == "localhost"
	if ip := net.ParseIP(host); ip != nil {
		loopback = ip.IsLoopback()
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && allowLoopback && loopback) {
		return "", Failure("client.https_required", "HTTPS is required; HTTP is permitted only for explicitly selected loopback development servers.", 2)
	}
	u.Host = strings.ToLower(u.Host)
	u.Path = ""
	return u.String(), nil
}

func NewClient(origin, token string, options Options) (*Client, error) {
	origin, err := NormalizeOrigin(origin, options.AllowLoopbackHTTP)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(token) == "" || strings.ContainsAny(token, "\r\n") {
		return nil, Failure("client.token_required", "An app token is required.", 3)
	}
	h := &http.Client{Timeout: 15 * time.Second}
	if options.HTTPClient != nil {
		*h = *options.HTTPClient
	}
	if h.Timeout == 0 || h.Timeout > 15*time.Second {
		h.Timeout = 15 * time.Second
	}
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{origin: origin, token: token, http: h}, nil
}

// call never retries: a failed write must be recovered using its original ID.
func (c *Client) call(ctx context.Context, tool string, args any, out any) error {
	id := c.nextID.Add(1)
	payload, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": "tools/call",
		"params": map[string]any{"name": tool, "arguments": args}})
	if err != nil {
		return Failure("client.invalid_arguments", "Invalid remote arguments.", 2)
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.origin+"/api/mcp", bytes.NewReader(payload))
	if err != nil {
		return Failure("client.request_failed", "Cannot prepare the remote request.", 5)
	}
	r.Header.Set("Authorization", "Bearer "+c.token)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	response, err := c.http.Do(r)
	if err != nil {
		return Failure("client.transport", "Could not obtain a response from Time & Bill.", 5)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		problem := Failure("client.authorization", "Time & Bill denied access. Check the app token and permissions.", 3)
		problem.definitive = true
		return problem
	}
	if response.StatusCode != http.StatusOK {
		return Failure("client.http_error", "Time & Bill returned an unsuccessful HTTP response.", 5)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(data) > maxResponseBytes {
		return Failure("client.invalid_response", "The remote response is incomplete or too large.", 5)
	}
	var envelope struct {
		Version string          `json:"jsonrpc"`
		ID      int64           `json:"id"`
		Error   json.RawMessage `json:"error"`
		Result  *struct {
			IsError *bool           `json:"isError"`
			Content json.RawMessage `json:"structuredContent"`
		} `json:"result"`
	}
	if json.Unmarshal(data, &envelope) != nil || envelope.Version != "2.0" || envelope.ID != id {
		return Failure("client.invalid_response", "Time & Bill returned an invalid protocol response.", 5)
	}
	if len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		return Failure("client.rpc_error", "Time & Bill could not process the remote request.", 5)
	}
	if envelope.Result == nil || envelope.Result.IsError == nil || len(envelope.Result.Content) == 0 || string(envelope.Result.Content) == "null" {
		return Failure("client.invalid_response", "Time & Bill returned an incomplete tool response.", 5)
	}
	if *envelope.Result.IsError {
		// Rejected writes may carry a durable receipt alongside the error.
		_ = json.Unmarshal(envelope.Result.Content, out)
		var problem struct {
			Code  string `json:"code"`
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if json.Unmarshal(envelope.Result.Content, &problem) != nil {
			return Failure("client.invalid_response", "Time & Bill returned an invalid error response.", 5)
		}
		code := problem.Error.Code
		if code == "" {
			code = problem.Code
		}
		if !safeCode(code) {
			code = "client.remote_error"
		}
		failure := remoteError(code)
		failure.definitive = true
		return failure
	}
	if json.Unmarshal(envelope.Result.Content, out) != nil {
		return Failure("client.invalid_response", "Time & Bill returned invalid tool data.", 5)
	}
	return nil
}

func safeCode(value string) bool {
	if value == "" || len(value) > 80 {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.' || r == '_') {
			return false
		}
	}
	return true
}
