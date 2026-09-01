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
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// MatchComparer is able to match and compare platforms to
// filter and sort platforms.
type MatchComparer interface {
	Matcher

	Less(specs.Platform, specs.Platform) bool
}

// Only returns a match comparer for a single platform using default
// resolution logic for the platform.
//
// Match and Less are independent: Match decides compatibility directly
// (see archMatch, osIdentity, featuresOK) instead of generating and
// searching every platform that could match, which would be unbounded for
// some variant schemes (see variant.go). Less is a separate,
// match-independent sort key biased towards the host's own OS and
// architecture. Callers wanting the best runnable match should filter with
// Match before ranking with Less, not sort first and scan for a match.
func Only(platform specs.Platform) MatchComparer {
	return &onlyComparer{
		platform: Normalize(platform),
	}
}

// onlyComparer implements the matching and ranking behavior of Only.
type onlyComparer struct {
	platform specs.Platform // normalized reference platform
}

// osIdentity reports whether p's OS name and OS version are ones
// c.platform could run. There's no "native vs. fallback OS" concept in Only
// (that would be for something like running Linux containers on Windows or
// FreeBSD), so an OS identity mismatch is an absolute disqualifier.
func (c *onlyComparer) osIdentity(p specs.Platform) bool {
	normalized := Normalize(p)
	if c.platform.OS != normalized.OS {
		return false
	}
	return osVersionMatch(c.platform.OS, c.platform.OSVersion, p.OSVersion)
}

// featuresOK reports whether p's OS features (ignoring win32k on Windows,
// which is missing on Nano Server) are a subset of c.platform's.
func (c *onlyComparer) featuresOK(p specs.Platform) bool {
	features := c.stripIgnoredFeatures(Normalize(p).OSFeatures)
	return osFeaturesSubset(features, c.platform.OSFeatures)
}

// featureOverlap returns how many of p's OS features are also declared by
// c.platform. Used only by Less, to rank two candidates that are otherwise
// tied: a candidate's extra features should only count in its favor to the
// extent the host actually declares them.
func (c *onlyComparer) featureOverlap(p specs.Platform) int {
	features := c.stripIgnoredFeatures(Normalize(p).OSFeatures)
	have := c.platform.OSFeatures
	n, j := 0, 0
	for _, f := range features {
		for j < len(have) && have[j] < f {
			j++
		}
		if j < len(have) && have[j] == f {
			n++
			j++
		}
	}
	return n
}

func (c *onlyComparer) stripIgnoredFeatures(features []string) []string {
	if c.platform.OS == "windows" {
		return stripWin32kFeature(features)
	}
	return features
}

// fallbackArch returns the recognized cross-architecture fallback for
// hostArch (386 for amd64, arm for arm64), or "" if hostArch has none. Both
// archMatch and archRank need this same pairing, one for compatibility and
// one for ranking, so it's kept in one place.
func fallbackArch(hostArch string) string {
	switch hostArch {
	case "amd64":
		return "386"
	case "arm64":
		return "arm"
	}
	return ""
}

// archMatch reports whether p's architecture and variant are compatible
// with c.platform's: either the same architecture with a compatible
// variant, or a recognized cross-architecture fallback (see fallbackArch)
// with a compatible variant.
//
// For amd64, floors at v1 (see numberedVariantMatch) and also matches 386
// via its fallback. For arm, floors at v5, so arm/v8 also matches arm/v7,
// arm/v6 and arm/v5 (and so on down the chain). For arm64, see
// arm64VariantMatch for its cross-generation offset. Any other architecture
// falls back to genericVariantMatch's prefix/number/suffix comparison (e.g.
// ppc64le's "powerN", riscv64's "rvaNNu64").
func (c *onlyComparer) archMatch(p specs.Platform) bool {
	normalized := Normalize(p)
	if normalized.Architecture == c.platform.Architecture {
		switch c.platform.Architecture {
		case "amd64":
			return numberedVariantMatch(c.platform.Variant, normalized.Variant, "v1")
		case "arm":
			return numberedVariantMatch(c.platform.Variant, normalized.Variant, "v5")
		case "arm64":
			return arm64VariantMatch(c.platform.Variant, normalized.Variant)
		default:
			return genericVariantMatch(c.platform.Variant, normalized.Variant)
		}
	}
	if normalized.Architecture != fallbackArch(c.platform.Architecture) {
		return false
	}

	switch c.platform.Architecture {
	case "amd64":
		return true // 386 has no variant to check; it's a match-or-nothing fallback.
	case "arm64":
		return numberedVariantMatch("v8", normalized.Variant, "v5")
	}
	return false
}

func (c *onlyComparer) Match(p specs.Platform) bool {
	return c.osIdentity(p) && c.featuresOK(p) && c.archMatch(p)
}

// archRank returns how preferable arch is, for a host declaring
// hostArch: 2 for the host's own architecture, 1 for its recognized
// cross-architecture fallback (see fallbackArch), or 0 otherwise (ties are
// broken alphabetically by Less). Unlike archMatch, this doesn't check
// variant compatibility at all — it's a plain preference between
// architecture names, used only for ranking.
func archRank(hostArch, arch string) int {
	switch {
	case arch == hostArch:
		return 2
	case arch != "" && arch == fallbackArch(hostArch):
		return 1
	default:
		return 0
	}
}

func (c *onlyComparer) Less(p1, p2 specs.Platform) bool {
	n1, n2 := Normalize(p1), Normalize(p2)

	// Prefer the host's own OS over any other; otherwise alphabetically.
	native1, native2 := n1.OS == c.platform.OS, n2.OS == c.platform.OS
	if native1 != native2 {
		return native1
	}
	if n1.OS != n2.OS {
		return naturalLess(n1.OS, n2.OS)
	}

	// Prefer the host's own architecture, then its recognized fallback
	// architecture, then alphabetically.
	a1, a2 := archRank(c.platform.Architecture, n1.Architecture), archRank(c.platform.Architecture, n2.Architecture)
	if a1 != a2 {
		return a1 > a2
	}
	if n1.Architecture != n2.Architecture {
		return naturalLess(n1.Architecture, n2.Architecture)
	}

	// Then variant and OS version, newest/highest first.
	if n1.Variant != n2.Variant {
		return naturalLess(n2.Variant, n1.Variant)
	}
	if p1.OSVersion != p2.OSVersion {
		return naturalLess(p2.OSVersion, p1.OSVersion)
	}

	// Tied on everything above: prefer the one whose OS features overlap
	// more with what the host actually declares.
	return c.featureOverlap(p1) > c.featureOverlap(p2)
}

// OnlyOS returns a match comparer that matches only platforms with the same
// OS, OS version, and OS features, regardless of architecture. When comparing,
// it always ranks the best architecture match highest using the default
// platform resolution logic.
func OnlyOS(platform specs.Platform) MatchComparer {
	normalized := Normalize(platform)
	return onlyOSComparer{
		platform: normalized,
		archOrder: orderedPlatformComparer{
			matchers: []Matcher{NewMatcher(normalized)},
		},
	}
}

type onlyOSComparer struct {
	platform  specs.Platform
	archOrder orderedPlatformComparer
}

func (c onlyOSComparer) matchOS(platform specs.Platform) bool {
	normalized := Normalize(platform)
	if c.platform.OS != normalized.OS {
		return false
	}
	if !osVersionMatch(c.platform.OS, c.platform.OSVersion, platform.OSVersion) {
		return false
	}
	return osFeaturesSubset(normalized.OSFeatures, c.platform.OSFeatures)
}

func (c onlyOSComparer) Match(platform specs.Platform) bool {
	return c.matchOS(platform)
}

func (c onlyOSComparer) Less(p1, p2 specs.Platform) bool {
	p1m := c.matchOS(p1)
	p2m := c.matchOS(p2)
	if p1m && !p2m {
		return true
	}
	if !p1m {
		return false
	}
	// Both match — rank by architecture preference
	return c.archOrder.Less(p1, p2)
}

// OnlyStrict returns a match comparer for a single platform.
//
// Unlike Only, OnlyStrict does not match sub platforms.
// So, "arm/vN" will not match "arm/vM" where M < N,
// and "amd64" will not also match "386".
//
// OnlyStrict matches non-canonical forms.
// So, "arm64" matches "arm/64/v8".
func OnlyStrict(platform specs.Platform) MatchComparer {
	return Ordered(Normalize(platform))
}

// Ordered returns a platform MatchComparer which matches any of the platforms
// but orders them in order they are provided.
func Ordered(platforms ...specs.Platform) MatchComparer {
	matchers := make([]Matcher, len(platforms))
	for i := range platforms {
		matchers[i] = NewMatcher(platforms[i])
	}
	return orderedPlatformComparer{
		matchers: matchers,
	}
}

// Any returns a platform MatchComparer which matches any of the platforms
// with no preference for ordering.
func Any(platforms ...specs.Platform) MatchComparer {
	matchers := make([]Matcher, len(platforms))
	for i := range platforms {
		matchers[i] = NewMatcher(platforms[i])
	}
	return anyPlatformComparer{
		matchers: matchers,
	}
}

// All is a platform MatchComparer which matches all platforms
// with preference for ordering.
var All MatchComparer = allPlatformComparer{}

type orderedPlatformComparer struct {
	matchers []Matcher
}

func (c orderedPlatformComparer) Match(platform specs.Platform) bool {
	for _, m := range c.matchers {
		if m.Match(platform) {
			return true
		}
	}
	return false
}

func (c orderedPlatformComparer) Less(p1 specs.Platform, p2 specs.Platform) bool {
	for _, m := range c.matchers {
		p1m := m.Match(p1)
		p2m := m.Match(p2)
		if p1m && !p2m {
			return true
		}
		if p1m || p2m {
			if p1m && p2m {
				// Prefer one with most matching features
				if len(p1.OSFeatures) != len(p2.OSFeatures) {
					return len(p1.OSFeatures) > len(p2.OSFeatures)
				}
			}
			return false
		}
	}
	if len(p1.OSFeatures) > 0 || len(p2.OSFeatures) > 0 {
		p1.OSFeatures = nil
		p2.OSFeatures = nil
		return c.Less(p1, p2)
	}
	return false
}

type anyPlatformComparer struct {
	matchers []Matcher
}

func (c anyPlatformComparer) Match(platform specs.Platform) bool {
	for _, m := range c.matchers {
		if m.Match(platform) {
			return true
		}
	}
	return false
}

func (c anyPlatformComparer) Less(p1, p2 specs.Platform) bool {
	var p1m, p2m bool
	for _, m := range c.matchers {
		if !p1m && m.Match(p1) {
			p1m = true
		}
		if !p2m && m.Match(p2) {
			p2m = true
		}
		if p1m && p2m {
			if len(p1.OSFeatures) != len(p2.OSFeatures) {
				return len(p1.OSFeatures) > len(p2.OSFeatures)
			}
			break
		}
	}

	// If neither match and has features, strip features and compare
	if !p1m && !p2m && (len(p1.OSFeatures) > 0 || len(p2.OSFeatures) > 0) {
		p1.OSFeatures = nil
		p2.OSFeatures = nil
		return c.Less(p1, p2)
	}

	// If one matches, and the other does, sort match first
	return p1m && !p2m
}

type allPlatformComparer struct{}

func (allPlatformComparer) Match(specs.Platform) bool {
	return true
}

func (allPlatformComparer) Less(specs.Platform, specs.Platform) bool {
	return false
}
