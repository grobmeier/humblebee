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
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestNativeCredentialStoreRoundTrip(t *testing.T) {
	if os.Getenv("HUMBLEBEE_NATIVE_KEYRING_TEST") != "1" {
		t.Skip("Requires an explicitly approved native-vault or disposable CI session")
	}
	ref, secret := "acceptance-"+uuid.NewString(), uuid.NewString()
	vault := NativeCredentials{}
	t.Cleanup(func() {
		if err := vault.Delete(ref); err != nil {
			t.Error("Cannot clean up the isolated acceptance credential")
		}
	})
	if err := vault.Set(ref, secret); err != nil {
		t.Fatal("Cannot write an isolated acceptance credential")
	}
	value, err := vault.Get(ref)
	if err != nil || value != secret {
		t.Fatal("Native credential round-trip failed")
	}
	if err := vault.Delete(ref); err != nil {
		t.Fatal("Cannot delete the isolated acceptance credential")
	}
	if _, err := vault.Get(ref); err == nil {
		t.Fatal("Deleted acceptance credential remains readable")
	}
}
