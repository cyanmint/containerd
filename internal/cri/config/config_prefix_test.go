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

package config

import (
	"testing"

	"github.com/cyanmint/containerd/v2/defaults"
)

func TestApplyPrefixToRuntimeDefaults(t *testing.T) {
	orig := defaults.PathPrefix
	defer func() { defaults.PathPrefix = orig }()

	const prefix = "/data/local/containerd"
	defaults.PathPrefix = prefix

	t.Run("default values are prefixed", func(t *testing.T) {
		cfg := DefaultRuntimeConfig()
		ApplyPrefixToRuntimeDefaults(&cfg)

		if len(cfg.CniConfig.NetworkPluginBinDirs) != 1 || cfg.CniConfig.NetworkPluginBinDirs[0] != prefix+defaultCNIBinDir {
			t.Errorf("NetworkPluginBinDirs: got %v, want [%s]", cfg.CniConfig.NetworkPluginBinDirs, prefix+defaultCNIBinDir)
		}
		if cfg.CniConfig.NetworkPluginConfDir != prefix+defaultCNIConfDir {
			t.Errorf("NetworkPluginConfDir: got %s, want %s", cfg.CniConfig.NetworkPluginConfDir, prefix+defaultCNIConfDir)
		}
		if len(cfg.CDISpecDirs) != 2 ||
			cfg.CDISpecDirs[0] != prefix+defaultCDISpecDir1 ||
			cfg.CDISpecDirs[1] != prefix+defaultCDISpecDir2 {
			t.Errorf("CDISpecDirs: got %v", cfg.CDISpecDirs)
		}
	})

	t.Run("explicitly-set values are not prefixed", func(t *testing.T) {
		cfg := DefaultRuntimeConfig()
		// Simulate a user explicitly setting these in the config file.
		cfg.CniConfig.NetworkPluginBinDirs = []string{"/custom/cni/bin"}
		cfg.CniConfig.NetworkPluginConfDir = "/custom/cni/net.d"
		cfg.CDISpecDirs = []string{"/custom/cdi"}

		ApplyPrefixToRuntimeDefaults(&cfg)

		if cfg.CniConfig.NetworkPluginBinDirs[0] != "/custom/cni/bin" {
			t.Errorf("explicit NetworkPluginBinDirs must not be prefixed, got %v", cfg.CniConfig.NetworkPluginBinDirs)
		}
		if cfg.CniConfig.NetworkPluginConfDir != "/custom/cni/net.d" {
			t.Errorf("explicit NetworkPluginConfDir must not be prefixed, got %s", cfg.CniConfig.NetworkPluginConfDir)
		}
		if cfg.CDISpecDirs[0] != "/custom/cdi" {
			t.Errorf("explicit CDISpecDirs must not be prefixed, got %v", cfg.CDISpecDirs)
		}
	})
}

func TestApplyPrefixToImageDefaults(t *testing.T) {
	orig := defaults.PathPrefix
	defer func() { defaults.PathPrefix = orig }()

	const prefix = "/data/local/containerd"
	defaults.PathPrefix = prefix

	t.Run("default registry path is prefixed", func(t *testing.T) {
		cfg := DefaultImageConfig()
		ApplyPrefixToImageDefaults(&cfg)

		want := prefix + defaultRegistryConfDir + ":" + prefix + defaultDockerCertsDir
		if cfg.Registry.ConfigPath != want {
			t.Errorf("ConfigPath: got %s, want %s", cfg.Registry.ConfigPath, want)
		}
	})

	t.Run("explicitly-set registry path is not prefixed", func(t *testing.T) {
		cfg := DefaultImageConfig()
		cfg.Registry.ConfigPath = "/custom/certs.d"

		ApplyPrefixToImageDefaults(&cfg)

		if cfg.Registry.ConfigPath != "/custom/certs.d" {
			t.Errorf("explicit ConfigPath must not be prefixed, got %s", cfg.Registry.ConfigPath)
		}
	})
}
