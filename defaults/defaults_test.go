/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package defaults

import (
	"path/filepath"
	"testing"
)

func TestPrefix(t *testing.T) {
	// Save and restore the global so parallel tests are unaffected.
	orig := PathPrefix
	defer func() { PathPrefix = orig }()

	t.Run("no prefix", func(t *testing.T) {
		PathPrefix = ""
		got := Prefix("/run/containerd")
		if got != "/run/containerd" {
			t.Errorf("expected /run/containerd, got %s", got)
		}
	})

	t.Run("with prefix", func(t *testing.T) {
		PathPrefix = "/data/local/containerd"
		got := Prefix("/run/containerd")
		want := filepath.Join("/data/local/containerd", "/run/containerd")
		if got != want {
			t.Errorf("expected %s, got %s", want, got)
		}
	})

	t.Run("Prefix always prepends PathPrefix to path", func(t *testing.T) {
		// Prefix() unconditionally prepends PathPrefix.  Callers are responsible
		// for NOT passing explicitly-configured (user-supplied) values through
		// Prefix() — they must use those values directly.  This test documents
		// the contract: Prefix() prepends, full stop.
		PathPrefix = "/data/local/containerd"
		got := Prefix("/run/containerd")
		want := filepath.Join("/data/local/containerd", "/run/containerd")
		if got != want {
			t.Errorf("expected %s, got %s", want, got)
		}
	})
}
