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
	"testing"
)

func TestParseVersionedVariant(t *testing.T) {
	for _, tc := range []struct {
		variant string
		prefix  string
		numbers []int
		suffix  string
		ok      bool
	}{
		{variant: "v3", prefix: "v", numbers: []int{3}, suffix: "", ok: true},
		{variant: "v8.5", prefix: "v", numbers: []int{8, 5}, suffix: "", ok: true},
		{variant: "power10", prefix: "power", numbers: []int{10}, suffix: "", ok: true},
		{variant: "rva23u64", prefix: "rva", numbers: []int{23}, suffix: "u64", ok: true},
		{variant: "10.0.17763.1", prefix: "", numbers: []int{10, 0, 17763, 1}, suffix: "", ok: true},
		{variant: "v8.3.1", prefix: "v", numbers: []int{8, 3, 1}, suffix: "", ok: true},
		{variant: "", ok: false},
		{variant: "custom", ok: false},
		{variant: "v99999999999999999999", ok: false}, // overflows int
	} {
		t.Run(tc.variant, func(t *testing.T) {
			p, ok := parseVersionedVariant(tc.variant)
			if ok != tc.ok {
				t.Fatalf("parseVersionedVariant(%q) ok = %v, want %v", tc.variant, ok, tc.ok)
			}
			if !tc.ok {
				return
			}
			if p.prefix != tc.prefix || !slices.Equal(p.numbers, tc.numbers) || p.suffix != tc.suffix {
				t.Fatalf("parseVersionedVariant(%q) = %+v, want {%q %v %q}", tc.variant, p, tc.prefix, tc.numbers, tc.suffix)
			}
		})
	}
}

// TestParseVersionedVariantStopsAtUndottedDigits ensures a digit run that
// isn't introduced by a literal "." is left in the opaque suffix rather
// than treated as another ranked version component. "." is the separator
// every dotted version scheme already uses; parsedVariant has no way to
// represent "components separated by something else", so a fixed
// identifier that happens to end in digits (like riscv64's "u64"/"s64"
// mode+width suffix) is compared as one exact-match string, never split.
func TestParseVersionedVariantStopsAtUndottedDigits(t *testing.T) {
	p, ok := parseVersionedVariant("rva20u64")
	if !ok {
		t.Fatal("expected rva20u64 to parse")
	}
	if !slices.Equal(p.numbers, []int{20}) {
		t.Fatalf("numbers = %v, want [20]", p.numbers)
	}
	if p.suffix != "u64" {
		t.Fatalf("suffix = %q, want %q", p.suffix, "u64")
	}
}

// TestNumberedVariantMatchHuge ensures compatibility between a host and
// image is computed directly, with no cost proportional to the magnitude of
// either variant's version number.
func TestNumberedVariantMatchHuge(t *testing.T) {
	if !numberedVariantMatch("v1000000000", "v999999999", "v1") {
		t.Fatal("expected a huge but lower image version to match")
	}
	if numberedVariantMatch("v5", "v6", "v1") {
		t.Fatal("expected higher image version to not match a lower host version")
	}
}

func TestGenericVariantMatch(t *testing.T) {
	for _, tc := range []struct {
		name   string
		host   string
		image  string
		wantOK bool
	}{
		{name: "exact", host: "power10", image: "power10", wantOK: true},
		{name: "lower", host: "power10", image: "power8", wantOK: true},
		{name: "higher", host: "power8", image: "power10", wantOK: false},
		{name: "bare image", host: "power10", image: "", wantOK: true},
		{name: "bare host, bare image", host: "", image: "", wantOK: true},
		{name: "bare host, versioned image", host: "", image: "power8", wantOK: false},
		{name: "different prefix", host: "power10", image: "rva20u64", wantOK: false},
		{name: "different suffix", host: "rva23u64", image: "rva20s64", wantOK: false},
		{name: "riscv lower", host: "rva23u64", image: "rva20u64", wantOK: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if ok := genericVariantMatch(tc.host, tc.image); ok != tc.wantOK {
				t.Fatalf("genericVariantMatch(%q, %q) = %v, want %v", tc.host, tc.image, ok, tc.wantOK)
			}
		})
	}
}

func TestArm64VariantMatch(t *testing.T) {
	for _, tc := range []struct {
		name   string
		host   string
		image  string
		wantOK bool
	}{
		{name: "v9.6 vs v8.9 (within v9.6+5 cap)", host: "v9.6", image: "v8.9", wantOK: true},
		{name: "v9.6 vs v8.10 (no such version, exceeds cap)", host: "v9.6", image: "v8.10", wantOK: false},
		{name: "v9 vs v8.5 (offset boundary)", host: "v9", image: "v8.5", wantOK: true},
		{name: "v9 vs v8.6 (past offset boundary)", host: "v9", image: "v8.6", wantOK: false},
		{name: "v8.1 vs v9 (lower major can't reach higher)", host: "v8.1", image: "v9", wantOK: false},
		{name: "v9.8 vs v9.0 (future minor, no table needed)", host: "v9.8", image: "v9.0", wantOK: true},
		{name: "v9.8 vs v8.9 (future minor, offset caps at .9)", host: "v9.8", image: "v8.9", wantOK: true},
		{name: "bare host is v8", host: "", image: "v8", wantOK: true},
		{name: "no third component", host: "v9.6", image: "v8.9.1", wantOK: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if ok := arm64VariantMatch(tc.host, tc.image); ok != tc.wantOK {
				t.Fatalf("arm64VariantMatch(%q, %q) = %v, want %v", tc.host, tc.image, ok, tc.wantOK)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b []int
		want int
	}{
		{[]int{8}, []int{8}, 0},
		{[]int{8}, []int{8, 0}, 0}, // missing trailing component treated as 0
		{[]int{8}, []int{8, 1}, -1},
		{[]int{8, 3}, []int{8, 9}, -1},
		{[]int{9, 0}, []int{8, 9}, 1}, // first component dominates
		{[]int{10, 0, 17763, 1}, []int{10, 0, 17763, 2}, -1},
	} {
		if got := compareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("compareVersions(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestNaturalLess(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"power9", "power10", true},
		{"power10", "power9", false},
		{"v8.5", "v8.10", true},
		{"v8.10", "v8.5", false},
		{"rva20u64", "rva23u64", true},
		{"rva23u64", "rva20u64", false},
		{"v9", "v9", false},
		{"v9", "v9.0", true}, // shorter string, otherwise identical, sorts first
		{"", "a", true},
		{"9", "09", true}, // numerically equal; fewer leading zeros sorts first
		{"09", "9", false},
		{"abc", "abd", true},
	} {
		t.Run(tc.a+"_"+tc.b, func(t *testing.T) {
			if got := naturalLess(tc.a, tc.b); got != tc.want {
				t.Errorf("naturalLess(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
