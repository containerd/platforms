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
	"fmt"
	"runtime"
	"sync"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sys/windows"
)

var (
	win32kOnce     sync.Once
	win32kFeatures []string
)

func detectWin32k() []string {
	win32kOnce.Do(func() {
		user32 := windows.NewLazySystemDLL("user32.dll")
		if err := user32.Load(); err == nil {
			win32kFeatures = []string{"win32k"}
		}
	})
	return win32kFeatures
}

// DefaultSpec returns the current platform's default platform specification.
func DefaultSpec() specs.Platform {
	major, minor, build := windows.RtlGetNtVersionNumbers()
	return specs.Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		OSVersion:    fmt.Sprintf("%d.%d.%d", major, minor, build),
		OSFeatures:   detectWin32k(),
		// The Variant field will be empty if arch != ARM.
		Variant: cpuVariant(),
	}
}

// Default returns the current platform's default platform specification.
func Default() MatchComparer {
	return &windowsMatchComparer{Matcher: NewMatcher(DefaultSpec())}
}
