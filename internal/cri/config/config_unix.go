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
	"github.com/cyanmint/containerd/v2/defaults"
	"github.com/pelletier/go-toml/v2"
)

// Default path constants used as "unset" sentinels in ApplyPrefix* functions.
const (
	defaultCNIBinDir       = "/opt/cni/bin"
	defaultCNIConfDir      = "/etc/cni/net.d"
	defaultCDISpecDir1     = "/etc/cdi"
	defaultCDISpecDir2     = "/var/run/cdi"
	defaultRegistryConfDir = "/etc/containerd/certs.d"
	defaultDockerCertsDir  = "/etc/docker/certs.d"
)

func defaultNetworkPluginBinDirs() []string {
	return []string{defaultCNIBinDir}
}

func DefaultImageConfig() ImageConfig {
	return ImageConfig{
		Snapshotter:                defaults.DefaultSnapshotter,
		DisableSnapshotAnnotations: true,
		MaxConcurrentDownloads:     3,
		Registry: Registry{
			ConfigPath: defaultRegistryConfDir + ":" + defaultDockerCertsDir,
		},
		ImageDecryption: ImageDecryption{
			KeyModel: KeyModelNode,
		},
		PinnedImages: map[string]string{
			"sandbox": DefaultSandboxImage,
		},
		ImagePullProgressTimeout: defaultImagePullProgressTimeoutDuration.String(),
		ImagePullWithSyncFs:      false,
		StatsCollectPeriod:       10,
	}
}

// DefaultRuntimeConfig returns default configurations of cri plugin.
func DefaultRuntimeConfig() RuntimeConfig {
	defaultRuncV2Opts := `
	# NoNewKeyring disables new keyring for the container.
	NoNewKeyring = false

	# ShimCgroup places the shim in a cgroup.
	ShimCgroup = ""

	# IoUid sets the I/O's pipes uid.
	IoUid = 0

	# IoGid sets the I/O's pipes gid.
	IoGid = 0

	# BinaryName is the binary name of the runc binary.
	BinaryName = ""

	# Root is the runc root directory.
	Root = ""

	# SystemdCgroup enables systemd cgroups.
	SystemdCgroup = false

	# CriuImagePath is the criu image path
	CriuImagePath = ""

	# CriuWorkPath is the criu work path.
	CriuWorkPath = ""
`
	var m map[string]interface{}
	toml.Unmarshal([]byte(defaultRuncV2Opts), &m)

	return RuntimeConfig{
		CniConfig: CniConfig{
			NetworkPluginBinDirs:       defaultNetworkPluginBinDirs(),
			NetworkPluginConfDir:       defaultCNIConfDir,
			NetworkPluginMaxConfNum:    1, // only one CNI plugin config file will be loaded
			NetworkPluginSetupSerially: false,
			NetworkPluginConfTemplate:  "",
			UseInternalLoopback:        false,
		},
		ContainerdConfig: ContainerdConfig{
			DefaultRuntimeName: "runc",
			Runtimes: map[string]Runtime{
				"runc": {
					Type:      "io.containerd.runc.v2",
					Options:   m,
					Sandboxer: string(ModePodSandbox),
				},
			},
		},
		EnableSelinux:                    false,
		SelinuxCategoryRange:             1024,
		MaxContainerLogLineSize:          16 * 1024,
		DisableProcMount:                 false,
		TolerateMissingHugetlbController: true,
		DisableHugetlbController:         true,
		IgnoreImageDefinedVolumes:        false,
		EnableCDI:                        func() *bool { v := true; return &v }(),
		CDISpecDirs:                      []string{defaultCDISpecDir1, defaultCDISpecDir2},
		DrainExecSyncIOTimeout:           "0s",
		EnableUnprivilegedPorts:          true,
		EnableUnprivilegedICMP:           true,
	}
}

// ApplyPrefixToRuntimeDefaults prepends defaults.PathPrefix to every CNI/CDI
// path in cfg that has not been overridden from its hard-coded default.
// Fields that differ from the known default (i.e. were set explicitly via the
// config file) are left unchanged.
// This function must be called inside the plugin InitFn, after config.Decode
// has merged any config-file values, so that defaults.PathPrefix is already set.
func ApplyPrefixToRuntimeDefaults(cfg *RuntimeConfig) {
	if defaults.PathPrefix == "" {
		return
	}
	// CNI bin dirs: only rewrite if still exactly the single-element default slice.
	if len(cfg.CniConfig.NetworkPluginBinDirs) == 1 &&
		cfg.CniConfig.NetworkPluginBinDirs[0] == defaultCNIBinDir {
		cfg.CniConfig.NetworkPluginBinDirs = []string{defaults.Prefix(defaultCNIBinDir)}
	}
	// CNI conf dir
	if cfg.CniConfig.NetworkPluginConfDir == defaultCNIConfDir {
		cfg.CniConfig.NetworkPluginConfDir = defaults.Prefix(defaultCNIConfDir)
	}
	// CDI spec dirs: only rewrite if still exactly the two-element default slice.
	if len(cfg.CDISpecDirs) == 2 &&
		cfg.CDISpecDirs[0] == defaultCDISpecDir1 &&
		cfg.CDISpecDirs[1] == defaultCDISpecDir2 {
		cfg.CDISpecDirs = []string{
			defaults.Prefix(defaultCDISpecDir1),
			defaults.Prefix(defaultCDISpecDir2),
		}
	}
}

// ApplyPrefixToImageDefaults prepends defaults.PathPrefix to the registry
// ConfigPath in cfg if it has not been overridden from its hard-coded default.
// This function must be called inside the plugin InitFn, after config.Decode
// has merged any config-file values.
func ApplyPrefixToImageDefaults(cfg *ImageConfig) {
	if defaults.PathPrefix == "" {
		return
	}
	if cfg.Registry.ConfigPath == defaultRegistryConfDir+":"+defaultDockerCertsDir {
		cfg.Registry.ConfigPath = defaults.Prefix(defaultRegistryConfDir) + ":" + defaults.Prefix(defaultDockerCertsDir)
	}
}
