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
	"path"
	"reflect"
	"runtime"
	"testing"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestParseSelector(t *testing.T) {
	var (
		defaultOS      = runtime.GOOS
		defaultArch    = runtime.GOARCH
		defaultVariant = ""
	)

	if defaultArch == "arm" && cpuVariant() != "v7" {
		defaultVariant = cpuVariant()
	}

	for _, testcase := range []struct {
		skip        bool
		input       string
		expected    specs.Platform
		matches     []specs.Platform
		formatted   string
		useV2Format bool
	}{
		// While wildcards are a valid use case for platform selection,
		// addressing these cases is outside the initial scope for this
		// package. When we do add platform wildcards, we should add in these
		// testcases to ensure that they are correctly represented.
		{
			skip:  true,
			input: "*",
			expected: specs.Platform{
				OS:           "*",
				Architecture: "*",
			},
			formatted:   "*/*",
			useV2Format: false,
		},
		{
			skip:  true,
			input: "linux/*",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "*",
			},
			formatted:   "linux/*",
			useV2Format: false,
		},
		{
			skip:  true,
			input: "*/arm64",
			expected: specs.Platform{
				OS:           "*",
				Architecture: "arm64",
			},
			matches: []specs.Platform{
				{
					OS:           "*",
					Architecture: "aarch64",
				},
				{
					OS:           "*",
					Architecture: "aarch64",
					Variant:      "v8",
				},
				{
					OS:           "*",
					Architecture: "arm64",
					Variant:      "v8",
				},
			},
			formatted:   "*/arm64",
			useV2Format: false,
		},
		{
			input: "linux/arm64",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "arm64",
			},
			matches: []specs.Platform{
				{
					OS:           "linux",
					Architecture: "aarch64",
				},
				{
					OS:           "linux",
					Architecture: "aarch64",
					Variant:      "v8",
				},
				{
					OS:           "linux",
					Architecture: "arm64",
					Variant:      "v8",
				},
			},
			formatted:   "linux/arm64",
			useV2Format: false,
		},
		{
			input: "linux/arm64/v8",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "arm64",
				Variant:      "v8",
			},
			matches: []specs.Platform{
				{
					OS:           "linux",
					Architecture: "aarch64",
				},
				{
					OS:           "linux",
					Architecture: "aarch64",
					Variant:      "v8",
				},
				{
					OS:           "linux",
					Architecture: "arm64",
				},
			},
			formatted:   "linux/arm64/v8",
			useV2Format: false,
		},
		{
			// NOTE(stevvooe): In this case, the consumer can assume this is v7
			// but we leave the variant blank. This will represent the vast
			// majority of arm images.
			input: "linux/arm",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "arm",
			},
			matches: []specs.Platform{
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "v7",
				},
				{
					OS:           "linux",
					Architecture: "armhf",
				},
				{
					OS:           "linux",
					Architecture: "arm",
					Variant:      "7",
				},
			},
			formatted:   "linux/arm",
			useV2Format: false,
		},
		{
			input: "linux/arm/v6",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v6",
			},
			matches: []specs.Platform{
				{
					OS:           "linux",
					Architecture: "armel",
				},
			},
			formatted:   "linux/arm/v6",
			useV2Format: false,
		},
		{
			input: "linux/arm/v7",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			matches: []specs.Platform{
				{
					OS:           "linux",
					Architecture: "arm",
				},
				{
					OS:           "linux",
					Architecture: "armhf",
				},
			},
			formatted:   "linux/arm/v7",
			useV2Format: false,
		},
		{
			input: "arm",
			expected: specs.Platform{
				OS:           defaultOS,
				Architecture: "arm",
			},
			formatted:   path.Join(defaultOS, "arm"),
			useV2Format: false,
		},
		{
			input: "armel",
			expected: specs.Platform{
				OS:           defaultOS,
				Architecture: "arm",
				Variant:      "v6",
			},
			formatted:   path.Join(defaultOS, "arm/v6"),
			useV2Format: false,
		},
		{
			input: "armhf",
			expected: specs.Platform{
				OS:           defaultOS,
				Architecture: "arm",
			},
			formatted:   path.Join(defaultOS, "arm"),
			useV2Format: false,
		},
		{
			input: "Aarch64",
			expected: specs.Platform{
				OS:           defaultOS,
				Architecture: "arm64",
			},
			formatted:   path.Join(defaultOS, "arm64"),
			useV2Format: false,
		},
		{
			input: "x86_64",
			expected: specs.Platform{
				OS:           defaultOS,
				Architecture: "amd64",
			},
			formatted:   path.Join(defaultOS, "amd64"),
			useV2Format: false,
		},
		{
			input: "Linux/x86_64",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			formatted:   "linux/amd64",
			useV2Format: false,
		},
		{
			input: "i386",
			expected: specs.Platform{
				OS:           defaultOS,
				Architecture: "386",
			},
			formatted:   path.Join(defaultOS, "386"),
			useV2Format: false,
		},
		{
			input: "linux",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("linux", defaultArch, defaultVariant),
			useV2Format: false,
		},
		{
			input: "s390x",
			expected: specs.Platform{
				OS:           defaultOS,
				Architecture: "s390x",
			},
			formatted:   path.Join(defaultOS, "s390x"),
			useV2Format: false,
		},
		{
			input: "linux/s390x",
			expected: specs.Platform{
				OS:           "linux",
				Architecture: "s390x",
			},
			formatted:   "linux/s390x",
			useV2Format: false,
		},
		{
			input: "macOS",
			expected: specs.Platform{
				OS:           "darwin",
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("darwin", defaultArch, defaultVariant),
			useV2Format: false,
		},
		{
			input: "windows",
			expected: specs.Platform{
				OS:           "windows",
				OSVersion:    "",
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("windows", defaultArch, defaultVariant),
			useV2Format: false,
		},
		{
			input: "windows()",
			expected: specs.Platform{
				OS:           "windows",
				OSVersion:    "",
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("windows", defaultArch, defaultVariant),
			useV2Format: true,
		},
		{
			input: "windows(10.0.17763)",
			expected: specs.Platform{
				OS:           "windows",
				OSVersion:    "10.0.17763",
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("windows(10.0.17763)", defaultArch, defaultVariant),
			useV2Format: true,
		},
		{
			input: "windows(10.0.17763)/amd64",
			expected: specs.Platform{
				OS:           "windows",
				OSVersion:    "10.0.17763",
				Architecture: "amd64",
			},
			formatted:   "windows(10.0.17763)/amd64",
			useV2Format: true,
		},
		{
			input: "macos(Abcd.Efgh.123-4)/aarch64",
			expected: specs.Platform{
				OS:           "darwin",
				OSVersion:    "Abcd.Efgh.123-4",
				Architecture: "arm64",
			},
			formatted:   "darwin(Abcd.Efgh.123-4)/arm64",
			useV2Format: true,
		},
	} {
		t.Run(testcase.input, func(t *testing.T) {
			if testcase.skip {
				t.Skip("this case is not yet supported")
			}
			p, err := Parse(testcase.input)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(p, testcase.expected) {
				t.Fatalf("platform did not match expected: %#v != %#v", p, testcase.expected)
			}

			m := NewMatcher(p)

			// ensure that match works on the input to the output.
			if ok := m.Match(testcase.expected); !ok {
				t.Fatalf("expected specifier %q matches %#v", testcase.input, testcase.expected)
			}
			for _, mc := range testcase.matches {
				if ok := m.Match(mc); !ok {
					t.Fatalf("expected specifier %q matches %#v", testcase.input, mc)
				}
			}

			formatted := ""
			if testcase.useV2Format {
				formatted = FormatAll(p)
			} else {
				formatted = Format(p)
			}
			if formatted != testcase.formatted {
				t.Fatalf("unexpected format: %q != %q", formatted, testcase.formatted)
			}

			// re-parse the formatted output and ensure we are stable
			reparsed, err := Parse(formatted)
			if err != nil {
				t.Fatalf("error parsing formatted output: %v", err)
			}

			if testcase.useV2Format {
				if FormatAll(reparsed) != formatted {
					t.Fatalf("normalized output did not survive the round trip: %v != %v", FormatAll(reparsed), formatted)
				}
			} else {
				if Format(reparsed) != formatted {
					t.Fatalf("normalized output did not survive the round trip: %v != %v", Format(reparsed), formatted)
				}
			}
		})
	}
}

func TestParseSelectorInvalid(t *testing.T) {
	for _, testcase := range []struct {
		input string
	}{
		{
			input: "", // empty
		},
		{
			input: "/linux/arm", // leading slash
		},
		{
			input: "linux/arm/", // trailing slash
		},
		{
			input: "linux /arm", // spaces
		},
		{
			input: "linux/&arm", // invalid character
		},
		{
			input: "linux/arm/foo/bar", // too many components
		},
		{
			input: "amd64/windows(10.0.17763)/foo", // only first element accepts os[(osVersion)]
		},
		{
			input: "linux)()---()..../arm/foo",
		},
		{
			input: "../arm/foo",
		},
	} {
		t.Run(testcase.input, func(t *testing.T) {
			if _, err := Parse(testcase.input); err == nil {
				t.Fatalf("should have received an error")
			}
		})
	}
}

func FuzzPlatformsParse(f *testing.F) {
	f.Add("linux/amd64")
	f.Fuzz(func(t *testing.T, s string) {
		pf, err := Parse(s)
		if err != nil && (pf.OS != "" || pf.Architecture != "") {
			t.Errorf("either %+v or %+v must be nil", err, pf)
		}
	})
}
