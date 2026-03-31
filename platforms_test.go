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
	"strconv"
	"strings"
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
			input: "windows(10.0.17763+win32k)",
			expected: specs.Platform{
				OS:           "windows",
				OSVersion:    "10.0.17763",
				OSFeatures:   []string{"win32k"},
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("windows(10.0.17763+win32k)", defaultArch, defaultVariant),
			useV2Format: true,
		},
		{
			input: "linux(+gpu)",
			expected: specs.Platform{
				OS:           "linux",
				OSVersion:    "",
				OSFeatures:   []string{"gpu"},
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("linux(+gpu)", defaultArch, defaultVariant),
			useV2Format: true,
		},
		{
			input: "linux(+gpu+simd)",
			expected: specs.Platform{
				OS:           "linux",
				OSVersion:    "",
				OSFeatures:   []string{"gpu", "simd"},
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("linux(+gpu+simd)", defaultArch, defaultVariant),
			useV2Format: true,
		},
		{
			input: "linux(+unsorted+erofs)",
			expected: specs.Platform{
				OS:           "linux",
				OSVersion:    "",
				OSFeatures:   []string{"unsorted", "erofs"},
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("linux(+erofs+unsorted)", defaultArch, defaultVariant),
			useV2Format: true,
		},
		{
			input: "windows(10.0.17763%2Bbuild.42)",
			expected: specs.Platform{
				OS:           "windows",
				OSVersion:    "10.0.17763+build.42",
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("windows(10.0.17763%2Bbuild.42)", defaultArch, defaultVariant),
			useV2Format: true,
		},
		{
			input: "windows(10.0.17763%2Bbuild.42+win32k)",
			expected: specs.Platform{
				OS:           "windows",
				OSVersion:    "10.0.17763+build.42",
				OSFeatures:   []string{"win32k"},
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("windows(10.0.17763%2Bbuild.42+win32k)", defaultArch, defaultVariant),
			useV2Format: true,
		},
		{
			input: "windows(50%25done)",
			expected: specs.Platform{
				OS:           "windows",
				OSVersion:    "50%done",
				Architecture: defaultArch,
				Variant:      defaultVariant,
			},
			formatted:   path.Join("windows(50%25done)", defaultArch, defaultVariant),
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
			if testcase.useV2Format == false {
				formatted = Format(p)
			} else {
				formatted = FormatAll(p)
			}
			if formatted != testcase.formatted {
				t.Fatalf("unexpected format: %q != %q", formatted, testcase.formatted)
			}

			// re-parse the formatted output and ensure we are stable
			reparsed, err := Parse(formatted)
			if err != nil {
				t.Fatalf("error parsing formatted output: %v", err)
			}

			if testcase.useV2Format == false {
				if Format(reparsed) != formatted {
					t.Fatalf("normalized output did not survive the round trip: %v != %v", Format(reparsed), formatted)
				}
			} else {
				if FormatAll(reparsed) != formatted {
					t.Fatalf("normalized output did not survive the round trip: %v != %v", FormatAll(reparsed), formatted)
				}
			}
		})
	}
}

func TestFormatAllEncoding(t *testing.T) {
	for _, testcase := range []struct {
		platform specs.Platform
		expected string
	}{
		{
			platform: specs.Platform{OS: "windows", OSVersion: "10.0.17763+build.42", Architecture: "amd64"},
			expected: "windows(10.0.17763%2Bbuild.42)/amd64",
		},
		{
			platform: specs.Platform{OS: "windows", OSVersion: "10.0.17763+build.42", OSFeatures: []string{"win32k"}, Architecture: "amd64"},
			expected: "windows(10.0.17763%2Bbuild.42+win32k)/amd64",
		},
		{
			platform: specs.Platform{OS: "windows", OSVersion: "50%done", Architecture: "amd64"},
			expected: "windows(50%25done)/amd64",
		},
		{
			platform: specs.Platform{OS: "windows", OSVersion: "1.0(beta)", Architecture: "amd64"},
			expected: "windows(1.0%28beta%29)/amd64",
		},
		{
			platform: specs.Platform{OS: "windows", OSVersion: "a/b", Architecture: "amd64"},
			expected: "windows(a%2Fb)/amd64",
		},
		{
			// no special characters, no encoding needed
			platform: specs.Platform{OS: "windows", OSVersion: "10.0.17763", Architecture: "amd64"},
			expected: "windows(10.0.17763)/amd64",
		},
		{
			// feature with + in the name
			platform: specs.Platform{OS: "linux", OSFeatures: []string{"feat+v2"}, Architecture: "amd64"},
			expected: "linux(+feat%2Bv2)/amd64",
		},
		{
			// feature with % in the name
			platform: specs.Platform{OS: "linux", OSFeatures: []string{"100%gpu"}, Architecture: "amd64"},
			expected: "linux(+100%25gpu)/amd64",
		},
		{
			// version and feature both with special characters
			platform: specs.Platform{OS: "windows", OSVersion: "10.0+build", OSFeatures: []string{"feat+1"}, Architecture: "amd64"},
			expected: "windows(10.0%2Bbuild+feat%2B1)/amd64",
		},
	} {
		t.Run(testcase.expected, func(t *testing.T) {
			formatted := FormatAll(testcase.platform)
			if formatted != testcase.expected {
				t.Fatalf("unexpected format: %q != %q", formatted, testcase.expected)
			}

			// verify round-trip
			reparsed, err := Parse(formatted)
			if err != nil {
				t.Fatalf("error parsing formatted output: %v", err)
			}
			if reparsed.OSVersion != testcase.platform.OSVersion {
				t.Fatalf("OSVersion did not survive round trip: %q != %q", reparsed.OSVersion, testcase.platform.OSVersion)
			}
			if !reflect.DeepEqual(reparsed.OSFeatures, testcase.platform.OSFeatures) {
				t.Fatalf("OSFeatures did not survive round trip: %v != %v", reparsed.OSFeatures, testcase.platform.OSFeatures)
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
	} {
		t.Run(testcase.input, func(t *testing.T) {
			if _, err := Parse(testcase.input); err == nil {
				t.Fatalf("should have received an error")
			}
		})
	}
}

func TestFormatAllSkipsEmptyOSFeatures(t *testing.T) {
	p := specs.Platform{
		OS:           "linux",
		Architecture: "amd64",
		OSFeatures:   []string{"", "gpu", "", "simd"},
	}

	formatted := FormatAll(p)
	expected := "linux(+gpu+simd)/amd64"
	if formatted != expected {
		t.Fatalf("unexpected format: %q != %q", formatted, expected)
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

func BenchmarkParseOSOptions(b *testing.B) {
	benchmarks := []struct {
		doc   string
		input string
	}{
		{
			doc:   "valid windows version and feature",
			input: "windows(10.0.17763+win32k)/amd64",
		},
		{
			doc:   "valid but lengthy features",
			input: "linux(+" + strings.Repeat("+feature", maxFeatures) + ")/amd64",
		},
		{
			doc:   "exploding plus chain",
			input: "linux(" + strings.Repeat("+", 64*1024) + ")/amd64",
		},
		{
			doc:   "kernel config feature blob",
			input: "linux(+CONFIG_" + strings.Repeat("FOO=y_", 16*1024) + "BAR)/amd64",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for _, bm := range benchmarks {
		b.Run(bm.doc, func(b *testing.B) {
			for range b.N {
				_, _ = Parse(bm.input)
			}
		})
	}
}

func BenchmarkFormatAllOSFeatures(b *testing.B) {
	benchmarks := []struct {
		doc      string
		platform specs.Platform
	}{
		{
			doc: "plain linux amd64",
			platform: specs.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
		{
			doc: "windows version and feature",
			platform: specs.Platform{
				OS:           "windows",
				OSVersion:    "10.0.17763",
				OSFeatures:   []string{"win32k"},
				Architecture: "amd64",
			},
		},
		{
			doc: "valid but lengthy features",
			platform: specs.Platform{
				OS: "linux",
				OSFeatures: func() (out []string) {
					for range maxFeatures {
						out = append(out, "feature")
					}
					return out
				}(),
				Architecture: "amd64",
			},
		},
		{
			doc: "skips empty features",
			platform: specs.Platform{
				OS:           "linux",
				OSFeatures:   []string{"", "gpu", "", "simd"},
				Architecture: "amd64",
			},
		},
		{
			doc: "kernel config feature blob",
			platform: specs.Platform{
				OS:           "linux",
				OSFeatures:   []string{"CONFIG_" + strings.Repeat("FOO_", 16*1024) + "BAR"},
				Architecture: "amd64",
			},
		},
		{
			doc: "many kernel config features with empties",
			platform: specs.Platform{
				OS: "linux",
				OSFeatures: func() []string {
					n := 1024
					out := make([]string, n)
					for i := range out {
						if i%10 == 0 {
							out[i] = "" // simulate bad data
						} else {
							out[i] = "CONFIG_FOO_" + strconv.Itoa(i)
						}
					}
					return out
				}(),
				Architecture: "amd64",
			},
		},
		{
			doc: "too many features",
			platform: specs.Platform{
				OS:           "linux",
				OSFeatures:   make([]string, maxFeatures+1),
				Architecture: "amd64",
			},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for _, bm := range benchmarks {
		b.Run(bm.doc, func(b *testing.B) {
			for range b.N {
				_ = FormatAll(bm.platform)
			}
		})
	}
}
