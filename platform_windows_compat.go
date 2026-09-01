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

package platforms

import (
	"slices"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// windowsOSVersion is a wrapper for Windows version information
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724439(v=vs.85).aspx
type windowsOSVersion struct {
	Version      uint32
	MajorVersion uint8
	MinorVersion uint8
	Build        uint16
}

// Windows Client and Server build numbers.
//
// See:
// https://learn.microsoft.com/en-us/windows/release-health/release-information
// https://learn.microsoft.com/en-us/windows/release-health/windows-server-release-info
// https://learn.microsoft.com/en-us/windows/release-health/windows11-release-information
const (
	// rs5 (version 1809, codename "Redstone 5") corresponds to Windows Server
	// 2019 (ltsc2019), and Windows 10 (October 2018 Update).
	rs5 = 17763
	// ltsc2019 (Windows Server 2019) is an alias for [rs5].
	ltsc2019 = rs5

	// v21H2Server corresponds to Windows Server 2022 (ltsc2022).
	v21H2Server = 20348
	// ltsc2022 (Windows Server 2022) is an alias for [v21H2Server]
	ltsc2022 = v21H2Server

	// v22H2Win11 corresponds to Windows 11 (2022 Update).
	v22H2Win11 = 22621

	// v23H2 is the 23H2 release in the Windows Server annual channel.
	v23H2 = 25398

	// Windows Server 2025 build 26100
	v25H1Server = 26100
	ltsc2025    = v25H1Server
)

// List of stable ABI compliant ltsc releases
// Note: List must be sorted in ascending order
var compatLTSCReleases = []uint16{
	ltsc2022,
	ltsc2025,
}

// CheckHostAndContainerCompat checks if given host and container
// OS versions are compatible.
// It includes support for stable ABI compliant versions as well.
// Every release after WS 2022 will support the previous ltsc
// container image. Stable ABI is in preview mode for windows 11 client.
// Refer: https://learn.microsoft.com/en-us/virtualization/windowscontainers/deploy-containers/version-compatibility?tabs=windows-server-2022%2Cwindows-10#windows-server-host-os-compatibility
func checkWindowsHostAndContainerCompat(host, ctr windowsOSVersion) bool {
	// check major minor versions of host and guest
	if host.MajorVersion != ctr.MajorVersion ||
		host.MinorVersion != ctr.MinorVersion {
		return false
	}

	// If host is < WS 2022, exact version match is required
	if host.Build < ltsc2022 {
		return host.Build == ctr.Build
	}

	// Find the floor of the compatible container range. Per the Windows stable
	// ABI policy, every host from LTSC N up to (but not including) LTSC N+1 can
	// run containers from LTSC N-1 up to the host build.
	//
	// So we find the largest LTSC <= host.Build, then step one entry back to
	// get the floor. If host.Build is past the latest LTSC in the list
	// (e.g. a 26200 host, which is in the WS2025 generation), the floor is
	// still the previous LTSC (20348), not the latest LTSC itself.
	//
	// If host is the very first LTSC (or no entry matches, which is impossible
	// here since we already checked host.Build >= ltsc2022), use that LTSC as
	// the floor.
	var supportedLTSCRelease uint16 = ltsc2022
	for i := len(compatLTSCReleases) - 1; i >= 0; i-- {
		if host.Build >= compatLTSCReleases[i] {
			if i == 0 {
				supportedLTSCRelease = compatLTSCReleases[i]
			} else {
				supportedLTSCRelease = compatLTSCReleases[i-1]
			}
			break
		}
	}
	return supportedLTSCRelease <= ctr.Build && ctr.Build <= host.Build
}

// getWindowsOSVersion parses the "<major>.<minor>.<build>[.<revision>]" form
// of a Windows OS version (e.g. "10.0.17763.1"), using the same
// parseVersionedVariant that parses every other dotted/numbered variant in
// this package — only the major/minor/build field mapping and their bit
// widths (matching the real Windows version struct) are specific to
// Windows here. Anything from the revision field onward is ignored, same as
// the Windows API this mirrors.
func getWindowsOSVersion(osVersionPrefix string) windowsOSVersion {
	v, ok := parseVersionedVariant(osVersionPrefix)
	if !ok || v.prefix != "" || v.suffix != "" || len(v.numbers) < 3 {
		return windowsOSVersion{}
	}
	major, minor, build := v.numbers[0], v.numbers[1], v.numbers[2]
	// parseVersionedVariant only ever produces non-negative numbers (they
	// come from a run of ASCII digits), but check explicitly anyway so the
	// range check below is a complete bound, not just an upper one.
	if major < 0 || major > 0xff || minor < 0 || minor > 0xff || build < 0 || build > 0xffff {
		return windowsOSVersion{}
	}

	return windowsOSVersion{
		MajorVersion: uint8(major),  // #nosec G115 -- range-checked above
		MinorVersion: uint8(minor),  // #nosec G115 -- range-checked above
		Build:        uint16(build), // #nosec G115 -- range-checked above
	}
}

// windowsOSVersionMatch reports whether a container declaring OS version v
// can run on a Windows host declaring hostVersion, per the Windows Server
// stable-ABI compatibility rules (see checkWindowsHostAndContainerCompat). A
// missing version on either side means "don't care", and always matches.
func windowsOSVersionMatch(hostVersion, v string) bool {
	host := getWindowsOSVersion(hostVersion)
	if host == (windowsOSVersion{}) || v == "" {
		return true
	}
	return checkWindowsHostAndContainerCompat(host, getWindowsOSVersion(v))
}

type windowsMatchComparer struct {
	Matcher
}

func (c *windowsMatchComparer) Less(p1, p2 specs.Platform) bool {
	m1, m2 := c.Match(p1), c.Match(p2)
	if m1 && m2 {
		return p1.OSVersion > p2.OSVersion
	}
	return m1 && !m2
}

// stripWin32kFeature removes "win32k" from features, if present, since it's
// missing on Nano Server and so is ignored for matching purposes. Returns
// features unmodified if "win32k" isn't present.
func stripWin32kFeature(features []string) []string {
	if i := slices.Index(features, "win32k"); i >= 0 {
		return slices.Delete(slices.Clone(features), i, i+1)
	}
	return features
}
