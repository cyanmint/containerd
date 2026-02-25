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

package oci

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func TestWithImageConfigNoEnv(t *testing.T) {
	t.Parallel()
	var (
		s   Spec
		c   = containers.Container{ID: t.Name()}
		ctx = namespaces.WithNamespace(context.Background(), "test")
	)

	err := populateDefaultUnixSpec(ctx, &s, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	// test hack: we don't want to test the WithAdditionalGIDs portion of the image config code
	s.Windows = &specs.Windows{}

	img, err := newFakeImage(ocispec.Image{
		Config: ocispec.ImageConfig{
			Entrypoint: []string{"create", "--namespace=test"},
			Cmd:        []string{"", "--debug"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	opts := []SpecOpts{
		WithImageConfigArgs(img, []string{"--boo", "bar"}),
	}

	// verify that if an image has no environment that we get a default Unix path
	expectedEnv := []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}

	for _, opt := range opts {
		if err := opt(nil, nil, nil, &s); err != nil {
			t.Fatal(err)
		}
	}

	if err := assertEqualsStringArrays(s.Process.Env, expectedEnv); err != nil {
		t.Fatal(err)
	}
}

func TestWithUmask_SetsUmaskOnEmptySpec(t *testing.T) {
	t.Parallel()
	var s Spec

	if err := WithUmask(0o027)(nil, nil, nil, &s); err != nil {
		t.Fatalf("WithUmask returned error: %v", err)
	}
	if s.Process == nil {
		t.Fatalf("expected Process to be initialized")
	}
	if s.Process.User.Umask == nil {
		t.Fatalf("expected Umask to be set")
	}
	if *s.Process.User.Umask != 0o027 {
		t.Fatalf("unexpected umask: got %03O, want %03O", *s.Process.User.Umask, 0o027)
	}
}

func TestWithUmask_WithDefaultSpec(t *testing.T) {
	t.Parallel()
	var s Spec
	c := containers.Container{ID: t.Name()}
	ctx := namespaces.WithNamespace(context.Background(), "test")

	// populate defaults first
	if err := populateDefaultUnixSpec(ctx, &s, c.ID); err != nil {
		t.Fatalf("populateDefaultUnixSpec error: %v", err)
	}

	// apply umask
	if err := WithUmask(0o077)(ctx, nil, &c, &s); err != nil {
		t.Fatalf("WithUmask returned error: %v", err)
	}

	if s.Process == nil || s.Process.User.Umask == nil {
		t.Fatalf("expected umask to be set on spec")
	}
	if *s.Process.User.Umask != 0o077 {
		t.Fatalf("unexpected umask: got %03O, want %03O", *s.Process.User.Umask, 0o077)
	}
}

// TestWithHostFileSkipsWhenMissing verifies that withHostFile silently skips the
// bind-mount when the source file does not exist on the host (e.g. Android/embedded
// systems that lack /etc/resolv.conf, /etc/hosts, or /etc/localtime).
func TestWithHostFileSkipsWhenMissing(t *testing.T) {
	t.Parallel()
	missingPath := filepath.Join(t.TempDir(), "nonexistent")
	var s Spec
	opt := withHostFile(missingPath, "/etc/test")
	if err := opt(nil, nil, nil, &s); err != nil {
		t.Fatalf("expected no error for missing host file, got: %v", err)
	}
	for _, m := range s.Mounts {
		if m.Destination == "/etc/test" {
			t.Errorf("expected no mount to be added for missing source %q, but found one", missingPath)
		}
	}
}

// TestWithHostFileAddsWhenPresent verifies that withHostFile adds the bind-mount
// when the source file exists on the host.
func TestWithHostFileAddsWhenPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "testfile")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	var s Spec
	opt := withHostFile(src, "/etc/test")
	if err := opt(nil, nil, nil, &s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, m := range s.Mounts {
		if m.Destination == "/etc/test" && m.Source == src {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected bind-mount %q -> /etc/test to be added but it was not", src)
	}
}
