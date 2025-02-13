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
	"testing"
)

// Test the platform compatibility of the different OS Versions
func Test_PlatformCompat(t *testing.T) {
	for _, tc := range []struct {
		hostOS    uint16
		ctrOS     uint16
		shouldRun bool
	}{
		{
			hostOS:    ltsc2019,
			ctrOS:     ltsc2019,
			shouldRun: true,
		},
		{
			hostOS:    ltsc2019,
			ctrOS:     ltsc2022,
			shouldRun: false,
		},
		{
			hostOS:    ltsc2022,
			ctrOS:     ltsc2019,
			shouldRun: false,
		},
		{
			hostOS:    ltsc2022,
			ctrOS:     ltsc2022,
			shouldRun: true,
		},
		{
			hostOS:    v22H2Win11,
			ctrOS:     ltsc2019,
			shouldRun: false,
		},
		{
			hostOS:    v22H2Win11,
			ctrOS:     ltsc2022,
			shouldRun: true,
		},
		{
			hostOS:    v23H2,
			ctrOS:     ltsc2019,
			shouldRun: false,
		},
		{
			hostOS:    v23H2,
			ctrOS:     ltsc2022,
			shouldRun: true,
		},
		{
			hostOS:    ltsc2025,
			ctrOS:     ltsc2022,
			shouldRun: true,
		},
		{
			hostOS:    ltsc2022,
			ctrOS:     ltsc2025,
			shouldRun: false,
		},
		{
			hostOS:    ltsc2022,
			ctrOS:     v22H2Win11,
			shouldRun: false,
		},
		{
			hostOS:    ltsc2025,
			ctrOS:     v22H2Win11,
			shouldRun: true,
		},
	} {
		t.Run(fmt.Sprintf("Host_%d_Ctr_%d", tc.hostOS, tc.ctrOS), func(t *testing.T) {
			hostOSVersion := windowsOSVersion{
				MajorVersion: 10,
				MinorVersion: 0,
				Build:        tc.hostOS,
			}
			ctrOSVersion := windowsOSVersion{
				MajorVersion: 10,
				MinorVersion: 0,
				Build:        tc.ctrOS,
			}
			if checkWindowsHostAndContainerCompat(hostOSVersion, ctrOSVersion) != tc.shouldRun {
				var expectedResultStr string
				if !tc.shouldRun {
					expectedResultStr = " NOT"
				}
				t.Fatalf("host %v should%s be able to run guest %v", tc.hostOS, expectedResultStr, tc.ctrOS)
			}
		})
	}
}
