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
	"strconv"
	"strings"
)

// parsedVariant is a variant string split around its dotted sequence of
// numbers, so two variants can be compared component-wise once their
// prefix/suffix are confirmed to match.
type parsedVariant struct {
	prefix  string
	numbers []int
	suffix  string
}

// parseVersionedVariant splits a variant into a leading non-numeric prefix,
// its leading run of digits, any further "."-separated runs of digits
// immediately following it, and whatever fixed text remains after that.
// The numbers are treated as the part of the variant that encodes a linear
// compatibility ordering; the prefix and suffix around them are held fixed.
// This covers every known OCI CPU variant scheme, and OS versions, without
// naming any of them:
//
//	"v3"           -> prefix "v",     numbers [3],             suffix ""
//	"v8.3"         -> prefix "v",     numbers [8, 3],          suffix ""
//	"power10"      -> prefix "power", numbers [10],            suffix ""
//	"rva23u64"     -> prefix "rva",   numbers [23],            suffix "u64"
//	"10.0.17763.1" -> prefix "",      numbers [10, 0, 17763, 1], suffix ""
//
// A "." is only ever treated as separating two numbers, never as starting
// the suffix: "rva23u64" stops at "u64" (not a digit after the — nonexistent
// — dot) without mistaking the "64" in "u64" for a second component.
func parseVersionedVariant(variant string) (parsedVariant, bool) {
	i := 0
	for i < len(variant) && (variant[i] < '0' || variant[i] > '9') {
		i++
	}
	prefix, rest := variant[:i], variant[i:]

	var numbers []int
	for {
		j := 0
		for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
			j++
		}
		if j == 0 {
			break
		}
		n, err := strconv.Atoi(rest[:j])
		if err != nil {
			// Overflows int (absurdly long digit run): treat as
			// unversioned rather than risk surprising wraparound behavior.
			return parsedVariant{}, false
		}
		numbers = append(numbers, n)
		rest = rest[j:]

		afterDot, isDotted := strings.CutPrefix(rest, ".")
		if !isDotted || afterDot == "" || afterDot[0] < '0' || afterDot[0] > '9' {
			break
		}
		rest = afterDot
	}
	if len(numbers) == 0 {
		return parsedVariant{}, false
	}
	return parsedVariant{prefix: prefix, numbers: numbers, suffix: rest}, true
}

// compareVersions compares two numeric version tuples component-wise, most
// significant first, treating a missing trailing component as 0 (so [8] ==
// [8, 0]). It returns -1, 0, or 1 as a < b, a == b, or a > b.
func compareVersions(a, b []int) int {
	for i := 0; i < max(len(a), len(b)); i++ {
		var x, y int
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
	}
	return 0
}

// sameShapeNumbers parses hostVariant and imageVariant and, if they share
// the same prefix and suffix (differing only in their numbers), returns
// both their numeric components for comparison.
func sameShapeNumbers(hostVariant, imageVariant string) (hostNumbers, imageNumbers []int, ok bool) {
	h, hok := parseVersionedVariant(hostVariant)
	i, iok := parseVersionedVariant(imageVariant)
	if !hok || !iok || h.prefix != i.prefix || h.suffix != i.suffix {
		return nil, nil, false
	}
	return h.numbers, i.numbers, true
}

// numberedVariantMatch reports whether an image declaring imageVariant can
// run on a host declaring hostVariant, for architectures with a real
// minimum ("floor") version — amd64 and arm — where a bare/empty variant is
// that architecture's own canonical form for its floor version, rather than
// "no requirement".
func numberedVariantMatch(hostVariant, imageVariant, floorVariant string) bool {
	if hostVariant == "" {
		hostVariant = floorVariant
	}
	if imageVariant == "" {
		imageVariant = floorVariant
	}
	if imageVariant == hostVariant {
		return true
	}

	hostNumbers, imageNumbers, ok := sameShapeNumbers(hostVariant, imageVariant)
	if !ok {
		return false
	}
	floor, _ := parseVersionedVariant(floorVariant)
	return compareVersions(imageNumbers, floor.numbers) >= 0 && compareVersions(imageNumbers, hostNumbers) <= 0
}

// genericVariantMatch handles any architecture without dedicated version
// handling (e.g. ppc64le's "powerN", riscv64's "rvaNNu64"), purely by the
// shape of the variant string rather than by architecture name. An image
// declaring no variant at all is treated as carrying no version
// requirement, so it matches any host of the same architecture.
func genericVariantMatch(hostVariant, imageVariant string) bool {
	if imageVariant == hostVariant || imageVariant == "" {
		return true
	}

	hostNumbers, imageNumbers, ok := sameShapeNumbers(hostVariant, imageVariant)
	if !ok {
		return false
	}
	return compareVersions(imageNumbers, hostNumbers) <= 0
}

// arm64MaxV8Minor is the highest Armv8 minor version that will ever exist:
// Arm's own architecture documentation states that Armv9.N always carries
// the mandatory feature baseline of Armv8.(N+5), and that the Armv8 line
// stops being extended at .9 (all further baseline growth happens under the
// v9 line). This makes the offset a fixed constant rather than a table that
// needs an entry added for every new minor version.
const arm64MaxV8Minor = 9

// arm64Minor returns numbers[1] (the minor version), or 0 if it isn't
// specified (a bare "v8" means the same thing as "v8.0").
func arm64Minor(numbers []int) int {
	if len(numbers) > 1 {
		return numbers[1]
	}
	return 0
}

// arm64VariantMatch reports whether an image declaring imageVariant can run
// on an arm64 host declaring hostVariant. Parsing is the same generic
// parseVersionedVariant used everywhere else; only the cross-generation
// offset below (see arm64MaxV8Minor) is genuinely arm64-specific, computed
// arithmetically instead of from a table enumerating every known
// major/minor pair.
//
// For arm64/v9.x, also matches arm64/v9.{0..x-1} and arm64/v8.{0..x+5}.
// For arm64/v8.x, also matches arm64/v8.{0..x-1}.
func arm64VariantMatch(hostVariant, imageVariant string) bool {
	if hostVariant == "" {
		hostVariant = "v8"
	}
	if imageVariant == "" {
		imageVariant = "v8"
	}
	if imageVariant == hostVariant {
		return true
	}

	h, hok := parseVersionedVariant(hostVariant)
	i, iok := parseVersionedVariant(imageVariant)
	if !hok || !iok || h.prefix != "v" || i.prefix != "v" || h.suffix != "" || i.suffix != "" ||
		len(h.numbers) > 2 || len(i.numbers) > 2 {
		return false
	}
	hMajor, hMinor := h.numbers[0], arm64Minor(h.numbers)
	iMajor, iMinor := i.numbers[0], arm64Minor(i.numbers)

	if iMajor == hMajor {
		return iMinor <= hMinor
	}
	if hMajor == 9 && iMajor == 8 {
		return iMinor <= min(hMinor+5, arm64MaxV8Minor)
	}
	return false
}

// naturalLess reports whether a sorts before b in "natural" order: runs of
// ASCII digits are compared as numbers (so "9" < "10"), and everything else
// is compared as plain text. It has no knowledge of any particular variant
// or version scheme's shape — it's the same ordering a file manager uses to
// sort "power9" before "power10", or "v8.9" before "v8.10".
func naturalLess(a, b string) bool {
	for len(a) > 0 && len(b) > 0 {
		da, ra := splitLeadingDigits(a)
		db, rb := splitLeadingDigits(b)
		if da != "" && db != "" {
			na := strings.TrimLeft(da, "0")
			nb := strings.TrimLeft(db, "0")
			switch {
			case len(na) != len(nb):
				return len(na) < len(nb)
			case na != nb:
				return na < nb
			case len(da) != len(db):
				// Numerically equal (differing only in leading zeros):
				// fewer leading zeros sorts first, for a total order.
				return len(da) < len(db)
			}
			a, b = ra, rb
			continue
		}
		if a[0] != b[0] {
			return a[0] < b[0]
		}
		a, b = a[1:], b[1:]
	}
	return len(a) < len(b)
}

func splitLeadingDigits(s string) (digits, rest string) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return s[:i], s[i:]
}
