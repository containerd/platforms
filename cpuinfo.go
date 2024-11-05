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
	"runtime"
	"sync"

	"github.com/containerd/log"
	amd64variant "github.com/tonistiigi/go-archvariant"
)

// Present the instruction set architecture, eg: v7, v8 for ARM CPU,
// v3, v4 for AMD64 CPU.
// Don't use this value directly; call cpuVariant() instead.
var cpuVariantValue string

var cpuVariantOnce sync.Once

func cpuVariant() string {
	cpuVariantOnce.Do(func() {
		if isArmArch(runtime.GOARCH) {
			var err error
			cpuVariantValue, err = getArmCPUVariant()
			if err != nil {
				log.L.Errorf("Error getArmCPUVariant for OS %s: %v", runtime.GOOS, err)
			}
		}
	})
	return cpuVariantValue
}

func cpuVariantMaximum() string {
	cpuVariantOnce.Do(func() {
		if isArmArch(runtime.GOARCH) {
			var err error
			cpuVariantValue, err = getArmCPUVariant()
			if err != nil {
				log.L.Errorf("Error getArmCPUVariant for OS %s: %v", runtime.GOOS, err)
			}
		} else if isAmd64Arch(runtime.GOARCH) {
			var err error
			cpuVariantValue, err = getAmd64MicroArchLevel()
			if err != nil {
				log.L.Errorf("Error getAmd64MicroArchLevel for OS %s: %v", runtime.GOOS, err)
			}
		}
	})
	return cpuVariantValue
}

func getAmd64MicroArchLevel() (string, error) {
	return amd64variant.AMD64Variant(), nil
}
