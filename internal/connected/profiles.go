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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
)

type Credentials interface {
	Get(string) (string, error)
	Set(string, string) error
	Delete(string) error
}

type Profile struct {
	Name              string `json:"name"`
	Origin            string `json:"origin"`
	AccountIdentity   string `json:"accountIdentity"`
	Timezone          string `json:"timezone"`
	CredentialRef     string `json:"credentialRef,omitempty"`
	AllowLoopbackHTTP bool   `json:"allowLoopbackHttp,omitempty"`
}

type Session struct {
	Profile      Profile
	Capabilities Capabilities
	Client       *Client
}

type Store struct {
	dir         string
	credentials Credentials
}

type profileFile struct {
	Version  int                `json:"version"`
	Profiles map[string]Profile `json:"profiles"`
}

var profileName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

func credentialKey(p Profile) string {
	data, _ := json.Marshal([]string{p.Name, p.Origin, p.AccountIdentity, p.CredentialRef})
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func NewStore(dir string, credentials Credentials) *Store {
	return &Store{dir: dir, credentials: credentials}
}

func (s *Store) load() (profileFile, error) {
	file := profileFile{Version: 1, Profiles: map[string]Profile{}}
	f, err := os.Open(filepath.Join(s.dir, "profiles.json"))
	if errors.Is(err, os.ErrNotExist) {
		return file, nil
	}
	if err != nil {
		return file, Failure("client.profile_read", "Cannot read remote profiles.", 5)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if err != nil || len(data) > 1<<20 || json.Unmarshal(data, &file) != nil || file.Version != 1 || file.Profiles == nil {
		return file, Failure("client.profile_invalid", "Remote profile configuration is invalid.", 5)
	}
	for name, p := range file.Profiles {
		if !profileName.MatchString(name) || p.Name != name || p.AccountIdentity == "" || p.Timezone == "" {
			return file, Failure("client.profile_invalid", "Remote profile configuration is invalid.", 5)
		}
	}
	return file, nil
}

func (s *Store) lock() (func(), error) {
	if os.MkdirAll(s.dir, 0700) != nil {
		return nil, Failure("client.profile_write", "Cannot create remote configuration directory.", 5)
	}
	path := filepath.Join(s.dir, "profiles.lock")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, Failure("client.profile_busy", "Remote configuration is locked. Check for another HumbleBee process before removing a stale profiles.lock.", 5)
	}
	_ = f.Close()
	return func() { _ = os.Remove(path) }, nil
}

func writePrivateJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return Failure("client.state_write", "Cannot encode remote state.", 5)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".remote-*")
	if err != nil {
		return Failure("client.state_write", "Cannot save remote state.", 5)
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		return Failure("client.state_write", "Cannot save remote state.", 5)
	}
	return nil
}

func (s *Store) Connect(ctx context.Context, name, origin, token string, ephemeral, replace bool, options Options) (Profile, error) {
	if !profileName.MatchString(name) {
		return Profile{}, Failure("client.profile_name", "Use a profile name containing letters, digits, hyphens or underscores.", 2)
	}
	client, err := NewClient(origin, token, options)
	if err != nil {
		return Profile{}, err
	}
	caps, err := client.Capabilities(ctx)
	if err != nil {
		return Profile{}, err
	}
	unlock, err := s.lock()
	if err != nil {
		return Profile{}, err
	}
	defer unlock()
	file, err := s.load()
	if err != nil {
		return Profile{}, err
	}
	old, exists := file.Profiles[name]
	if exists && !replace {
		return Profile{}, Failure("client.profile_exists", "Profile already exists. Use --replace to confirm replacing its server/account binding.", 2)
	}
	p := Profile{Name: name, Origin: client.origin, AccountIdentity: caps.AccountIdentity, Timezone: caps.Timezone, AllowLoopbackHTTP: options.AllowLoopbackHTTP}
	if !ephemeral {
		p.CredentialRef = uuid.NewString()
		if s.credentials == nil || s.credentials.Set(credentialKey(p), token) != nil {
			return Profile{}, Failure("client.keyring_unavailable", "Cannot use secure credential storage. Use HUMBLEBEE_TAB_TOKEN with --ephemeral instead; no token was saved in a file.", 3)
		}
	}
	file.Profiles[name] = p
	if err := writePrivateJSON(filepath.Join(s.dir, "profiles.json"), file); err != nil {
		if p.CredentialRef != "" {
			_ = s.credentials.Delete(credentialKey(p))
		}
		return Profile{}, err
	}
	if exists && old.CredentialRef != "" && s.credentials != nil {
		if s.credentials.Delete(credentialKey(old)) != nil {
			return p, Failure("client.credential_cleanup", "Profile replaced, but the old credential could not be removed from secure storage.", 5)
		}
	}
	return p, nil
}

func (s *Store) Select(name string) (Profile, error) {
	file, err := s.load()
	if err != nil {
		return Profile{}, err
	}
	if name == "" && len(file.Profiles) == 1 {
		for _, p := range file.Profiles {
			return p, nil
		}
	}
	p, ok := file.Profiles[name]
	if !ok {
		return Profile{}, Failure("client.profile_required", "Select an existing remote profile with --profile. Connect first if none exists.", 2)
	}
	return p, nil
}

func (s *Store) Open(ctx context.Context, name, override string, options Options) (*Session, error) {
	p, err := s.Select(name)
	if err != nil {
		return nil, err
	}
	token := override
	if token == "" {
		if p.CredentialRef == "" || s.credentials == nil {
			return nil, Failure("client.token_required", "Provide HUMBLEBEE_TAB_TOKEN for this ephemeral profile.", 3)
		}
		token, err = s.credentials.Get(credentialKey(p))
		if err != nil {
			return nil, Failure("client.keyring_unavailable", "Cannot read the app token from secure credential storage.", 3)
		}
	}
	options.AllowLoopbackHTTP = p.AllowLoopbackHTTP
	c, err := NewClient(p.Origin, token, options)
	if err != nil {
		return nil, err
	}
	caps, err := c.Capabilities(ctx)
	if err != nil {
		return nil, err
	}
	if caps.AccountIdentity != p.AccountIdentity {
		return nil, Failure("client.account_changed", "Remote account identity changed. No operation was performed; reconnect explicitly.", 3)
	}
	p.Timezone = caps.Timezone
	return &Session{Profile: p, Capabilities: caps, Client: c}, nil
}

func (s *Store) Disconnect(name string) error {
	unlock, err := s.lock()
	if err != nil {
		return err
	}
	defer unlock()
	p, err := s.Select(name)
	if err != nil {
		return err
	}
	file, err := s.load()
	if err != nil {
		return err
	}
	if p.CredentialRef != "" && (s.credentials == nil || s.credentials.Delete(credentialKey(p)) != nil) {
		return Failure("client.keyring_unavailable", "Cannot remove the stored credential; profile retained so you can retry.", 3)
	}
	delete(file.Profiles, p.Name)
	return writePrivateJSON(filepath.Join(s.dir, "profiles.json"), file)
}
