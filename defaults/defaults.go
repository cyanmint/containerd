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

import "path/filepath"

// PathPrefix is an optional prefix prepended to all default paths.
// It is set once at process startup (e.g. from the --prefix flag) and is
// never changed afterwards.  Code that computes default paths should call
// Prefix() rather than hard-coding the base path.
var PathPrefix string

// Prefix prepends PathPrefix to path.  If PathPrefix is empty the original
// path is returned unchanged.  An explicitly-configured value (i.e. one not
// derived from a default) must never be passed through this function.
func Prefix(path string) string {
	if PathPrefix == "" {
		return path
	}
	return filepath.Join(PathPrefix, path)
}

const (
	// DefaultMaxRecvMsgSize defines the default maximum message size for
	// receiving protobufs passed over the GRPC API.
	DefaultMaxRecvMsgSize = 16 << 20
	// DefaultMaxSendMsgSize defines the default maximum message size for
	// sending protobufs passed over the GRPC API.
	DefaultMaxSendMsgSize = 16 << 20
	// DefaultRuntimeNSLabel defines the namespace label to check for the
	// default runtime
	DefaultRuntimeNSLabel = "containerd.io/defaults/runtime"
	// DefaultSnapshotterNSLabel defines the namespace label to check for the
	// default snapshotter
	DefaultSnapshotterNSLabel = "containerd.io/defaults/snapshotter"
	// DefaultSandboxerNSLabel defines the namespace label to check for the
	// default sandboxcr
	DefaultSandboxerNSLabel = "containerd.io/defaults/sandboxer"
	// DefaultSandboxer defines the default sandboxer to use for creating sandboxes.
	DefaultSandboxer = "shim"
)
