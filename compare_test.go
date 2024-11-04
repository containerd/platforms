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
	"testing"
)

func TestOnly(t *testing.T) {
	for _, tc := range []struct {
		platform string
		matches  map[bool][]string
	}{
		{
			platform: "linux/amd64",
			matches: map[bool][]string{
				true: {
					"linux/amd64",
					"linux/386",
				},
				false: {
					"linux/amd64/v2",
					"linux/arm/v7",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/amd64/v2",
			matches: map[bool][]string{
				true: {
					"linux/amd64",
					"linux/amd64/v1",
					"linux/amd64/v2",
					"linux/386",
				},
				false: {
					"linux/amd64/v3",
					"linux/amd64/v4",
					"linux/arm/v7",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/386",
			matches: map[bool][]string{
				true: {
					"linux/386",
				},
				false: {
					"linux/amd64",
					"linux/arm/v7",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "windows/amd64",
			matches: map[bool][]string{
				true: {
					"windows/amd64",
					"windows(10.0.17763)/amd64",
				},
				false: {
					"linux/amd64",
					"linux/arm/v7",
					"linux/arm64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v8",
			matches: map[bool][]string{
				true: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
				},
				false: {
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v7",
			matches: map[bool][]string{
				true: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
				},
				false: {
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm/v8",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v6",
			matches: map[bool][]string{
				true: {
					"linux/arm/v5",
					"linux/arm/v6",
				},
				false: {
					"linux/amd64",
					"linux/arm",
					"linux/arm/v4",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v5",
			matches: map[bool][]string{
				true: {
					"linux/arm/v5",
				},
				false: {
					"linux/amd64",
					"linux/arm",
					"linux/arm/v4",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v4",
			matches: map[bool][]string{
				true: {
					"linux/arm/v4",
				},
				false: {
					"linux/amd64",
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm64/v9.6",
			matches: map[bool][]string{
				true: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"linux/arm64/v8",
					"linux/arm64/v8.1",
					"linux/arm64/v8.2",
					"linux/arm64/v8.3",
					"linux/arm64/v8.4",
					"linux/arm64/v8.5",
					"linux/arm64/v8.6",
					"linux/arm64/v8.7",
					"linux/arm64/v8.8",
					"linux/arm64/v8.9",
					"linux/arm64/v9",
					"linux/arm64/v9.0",
					"linux/arm64/v9.1",
					"linux/arm64/v9.2",
					"linux/arm64/v9.3",
					"linux/arm64/v9.4",
					"linux/arm64/v9.5",
					"linux/arm64/v9.6",
				},
				false: {
					"linux/amd64",
					"linux/arm/v4",
					"windows/amd64",
					"windows/arm",
					"linux/arm64/v8.10", // there's no v8.10
				},
			},
		},
		{
			platform: "linux/arm64/v9",
			matches: map[bool][]string{
				true: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"linux/arm64/v8",
					"linux/arm64/v8.1",
					"linux/arm64/v8.2",
					"linux/arm64/v8.3",
					"linux/arm64/v8.4",
					"linux/arm64/v8.5",
					"linux/arm64/v9",
					"linux/arm64/v9.0",
				},
				false: {
					"linux/amd64",
					"linux/arm/v4",
					"windows/amd64",
					"windows/arm",
					"linux/arm64/v8.6",
				},
			},
		},
		{
			platform: "linux/arm64/v8.1",
			matches: map[bool][]string{
				true: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"linux/arm64/v8",
					"linux/arm64/v8.1",
				},
				false: {
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm/v9",
					"linux/arm64/v9",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm64",
			matches: map[bool][]string{
				true: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"linux/arm64/v8",
				},
				false: {
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm/v9",
					"linux/arm64/v9",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
	} {
		testcase := tc
		t.Run(testcase.platform, func(t *testing.T) {
			p, err := Parse(testcase.platform)
			if err != nil {
				t.Fatal(err)
			}
			m := Only(p)
			for shouldMatch, platforms := range testcase.matches {
				for _, matchPlatform := range platforms {
					mp, err := Parse(matchPlatform)
					if err != nil {
						t.Fatal(err)
					}
					if match := m.Match(mp); shouldMatch != match {
						t.Errorf("Only(%q).Match(%q) should return %v, but returns %v", testcase.platform, matchPlatform, shouldMatch, match)
					}
				}
			}
		})
	}
}

func TestOnlyStrict(t *testing.T) {
	for _, tc := range []struct {
		platform string
		matches  map[bool][]string
	}{
		{
			platform: "linux/amd64",
			matches: map[bool][]string{
				true: {
					"linux/amd64",
				},
				false: {
					"linux/386",
					"linux/arm/v7",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/386",
			matches: map[bool][]string{
				true: {
					"linux/386",
				},
				false: {
					"linux/amd64",
					"linux/arm/v7",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "windows/amd64",
			matches: map[bool][]string{
				true: {
					"windows/amd64",
					"windows(10.0.17763)/amd64",
				},
				false: {
					"linux/amd64",
					"linux/arm/v7",
					"linux/arm64",
					"windows/arm",
				},
			},
		},
		{
			platform: "windows(10.0.17763)/amd64",
			matches: map[bool][]string{
				true: {
					"windows/amd64",
					"windows(10.0.17763)/amd64",
				},
				false: {
					"linux/amd64",
					"linux/arm/v7",
					"linux/arm64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v8",
			matches: map[bool][]string{
				true: {
					"linux/arm/v8",
				},
				false: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v7",
			matches: map[bool][]string{
				true: {
					"linux/arm",
					"linux/arm/v7",
				},
				false: {
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm/v8",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v6",
			matches: map[bool][]string{
				true: {
					"linux/arm/v6",
				},
				false: {
					"linux/arm/v5",
					"linux/amd64",
					"linux/arm",
					"linux/arm/v4",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v5",
			matches: map[bool][]string{
				true: {
					"linux/arm/v5",
				},
				false: {
					"linux/amd64",
					"linux/arm",
					"linux/arm/v4",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm/v4",
			matches: map[bool][]string{
				true: {
					"linux/arm/v4",
				},
				false: {
					"linux/amd64",
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/arm64",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm64/v9.5",
			matches: map[bool][]string{
				true: {
					"linux/arm64/v9.5",
				},
				false: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm64/v8",
					"linux/arm64/v8.1",
					"linux/arm64/v9",
					"linux/arm64/v9.1",
					"linux/arm64/v9.4",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm64/v9",
			matches: map[bool][]string{
				true: {
					"linux/arm64/v9",
					"linux/arm64/v9.0",
				},
				false: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm/v9",
					"linux/arm64/v8",
					"linux/arm64/v8.1",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm64/v8.1",
			matches: map[bool][]string{
				true: {
					"linux/arm64/v8.1",
				},
				false: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm/v9",
					"linux/arm64/v8",
					"linux/arm64/v9",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
		{
			platform: "linux/arm64",
			matches: map[bool][]string{
				true: {
					"linux/arm64",
					"linux/arm64/v8",
				},
				false: {
					"linux/arm",
					"linux/arm/v5",
					"linux/arm/v6",
					"linux/arm/v7",
					"linux/arm/v8",
					"linux/amd64",
					"linux/arm/v4",
					"linux/arm/v9",
					"linux/arm64/v9",
					"windows/amd64",
					"windows/arm",
				},
			},
		},
	} {
		testcase := tc
		t.Run(testcase.platform, func(t *testing.T) {
			p, err := Parse(testcase.platform)
			if err != nil {
				t.Fatal(err)
			}
			m := OnlyStrict(p)
			for shouldMatch, platforms := range testcase.matches {
				for _, matchPlatform := range platforms {
					mp, err := Parse(matchPlatform)
					if err != nil {
						t.Fatal(err)
					}
					if match := m.Match(mp); shouldMatch != match {
						t.Errorf("OnlyStrict(%q).Match(%q) should return %v, but returns %v", testcase.platform, matchPlatform, shouldMatch, match)
					}
				}
			}
		})
	}
}
