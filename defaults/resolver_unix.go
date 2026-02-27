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
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
)

// ConfigureResolver overrides net.DefaultResolver to use the nameservers
// from Prefix("/etc/resolv.conf") when PathPrefix is set.  This is needed
// on devices where /etc/resolv.conf does not exist (e.g. Android/Termux):
// without it Go falls back to [::1]:53 and 127.0.0.1:53, which are often
// unreachable, causing all DNS lookups (and therefore image pulls) to fail.
// It is a no-op when PathPrefix is empty.
func ConfigureResolver() {
	if PathPrefix == "" {
		return
	}
	servers := readNameservers(Prefix("/etc/resolv.conf"))
	if len(servers) == 0 {
		return
	}
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			var lastErr error
			for _, s := range servers {
				conn, err := d.DialContext(ctx, "udp", net.JoinHostPort(s, "53"))
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}
			return nil, fmt.Errorf("all nameservers from %s failed (tried %v): %w",
				Prefix("/etc/resolv.conf"), servers, lastErr)
		},
	}
}

// readNameservers returns the nameserver IP addresses from a
// resolv.conf-style file.  Lines that begin with '#' are ignored.
func readNameservers(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var servers []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			servers = append(servers, fields[1])
		}
	}
	return servers
}
