//go:build !windows

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
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestReadNameservers(t *testing.T) {
	t.Run("valid entries", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "resolv.conf")
		if err := os.WriteFile(f, []byte("# comment\nnameserver 8.8.8.8\nnameserver 8.8.4.4\n"), 0644); err != nil {
			t.Fatal(err)
		}
		got := readNameservers(f)
		if len(got) != 2 || got[0] != "8.8.8.8" || got[1] != "8.8.4.4" {
			t.Errorf("unexpected servers: %v", got)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		got := readNameservers(filepath.Join(t.TempDir(), "no-such-file"))
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "resolv.conf")
		if err := os.WriteFile(f, []byte("# only comments\nsearch example.com\n"), 0644); err != nil {
			t.Fatal(err)
		}
		got := readNameservers(f)
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("ipv6 nameserver", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "resolv.conf")
		if err := os.WriteFile(f, []byte("nameserver 2001:4860:4860::8888\n"), 0644); err != nil {
			t.Fatal(err)
		}
		got := readNameservers(f)
		if len(got) != 1 || got[0] != "2001:4860:4860::8888" {
			t.Errorf("unexpected servers: %v", got)
		}
	})
}

func TestConfigureResolver(t *testing.T) {
	origPrefix := PathPrefix
	origResolver := net.DefaultResolver
	defer func() {
		PathPrefix = origPrefix
		net.DefaultResolver = origResolver
	}()

	t.Run("no-op when prefix empty", func(t *testing.T) {
		PathPrefix = ""
		net.DefaultResolver = origResolver
		ConfigureResolver()
		if net.DefaultResolver != origResolver {
			t.Error("resolver should not be changed when prefix is empty")
		}
	})

	t.Run("no-op when resolv.conf missing", func(t *testing.T) {
		PathPrefix = t.TempDir() // prefix dir has no etc/resolv.conf
		net.DefaultResolver = origResolver
		ConfigureResolver()
		if net.DefaultResolver != origResolver {
			t.Error("resolver should not be changed when PREFIX/etc/resolv.conf is missing")
		}
	})

	t.Run("overrides resolver when resolv.conf has nameservers", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "etc"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "etc", "resolv.conf"), []byte("nameserver 8.8.8.8\n"), 0644); err != nil {
			t.Fatal(err)
		}
		PathPrefix = dir
		net.DefaultResolver = origResolver
		ConfigureResolver()
		if net.DefaultResolver == origResolver {
			t.Error("resolver should have been replaced")
		}
	})
}
