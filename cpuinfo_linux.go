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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

// getMachineArch retrieves the machine architecture through system call
func getMachineArch() (string, error) {
	var uname unix.Utsname
	err := unix.Uname(&uname)
	if err != nil {
		return "", err
	}

	arch := string(uname.Machine[:bytes.IndexByte(uname.Machine[:], 0)])

	return arch, nil
}

// For Linux, the kernel has already detected the ABI, ISA and Features.
// So we don't need to access the ARM registers to detect platform information
// by ourselves. We can just parse these information from /proc/cpuinfo
func getCPUInfo(pattern string) (info string, err error) {

	cpuinfo, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", err
	}
	defer cpuinfo.Close()

	// Start to Parse the Cpuinfo line by line. For SMP SoC, we parse
	// the first core is enough.
	scanner := bufio.NewScanner(cpuinfo)
	for scanner.Scan() {
		newline := scanner.Text()
		list := strings.Split(newline, ":")

		if len(list) > 1 && strings.EqualFold(strings.TrimSpace(list[0]), pattern) {
			return strings.TrimSpace(list[1]), nil
		}
	}

	// Check whether the scanner encountered errors
	err = scanner.Err()
	if err != nil {
		return "", err
	}

	return "", fmt.Errorf("getCPUInfo for pattern %s: %w", pattern, errNotFound)
}

// getCPUVariantFromArch get CPU variant from arch through a system call
func getCPUVariantFromArch(arch string) (string, error) {

	var variant string

	arch = strings.ToLower(arch)

	if arch == "aarch64" {
		variant = "8"
	} else if arch[0:4] == "armv" && len(arch) >= 5 {
		// Valid arch format is in form of armvXx
		switch arch[3:5] {
		case "v8":
			variant = "8"
		case "v7":
			variant = "7"
		case "v6":
			variant = "6"
		case "v5":
			variant = "5"
		case "v4":
			variant = "4"
		case "v3":
			variant = "3"
		default:
			variant = "unknown"
		}
	} else {
		return "", fmt.Errorf("getCPUVariantFromArch invalid arch: %s, %w", arch, errInvalidArgument)
	}
	return variant, nil
}

// getArmCPUVariant returns cpu variant for ARM
// We first try reading "Cpu architecture" field from /proc/cpuinfo
// If we can't find it, then fall back using a system call
// This is to cover running ARM in emulated environment on x86 host as this field in /proc/cpuinfo
// was not present.
func getArmCPUVariant() (string, error) {
	variant, err := getCPUInfo("Cpu architecture")
	if err != nil {
		if errors.Is(err, errNotFound) {
			// Let's try getting CPU variant from machine architecture
			arch, err := getMachineArch()
			if err != nil {
				return "", fmt.Errorf("failure getting machine architecture: %v", err)
			}

			variant, err = getCPUVariantFromArch(arch)
			if err != nil {
				return "", fmt.Errorf("failure getting CPU variant from machine architecture: %v", err)
			}
		} else {
			return "", fmt.Errorf("failure getting CPU variant: %v", err)
		}
	}

	// handle edge case for Raspberry Pi ARMv6 devices (which due to a kernel quirk, report "CPU architecture: 7")
	// https://www.raspberrypi.org/forums/viewtopic.php?t=12614
	if runtime.GOARCH == "arm" && variant == "7" {
		model, err := getCPUInfo("model name")
		if err == nil && strings.HasPrefix(strings.ToLower(model), "armv6-compatible") {
			variant = "6"
		}
	}

	switch strings.ToLower(variant) {
	case "8", "aarch64":
		variant = "v8"
	case "7", "7m", "?(12)", "?(13)", "?(14)", "?(15)", "?(16)", "?(17)":
		variant = "v7"
	case "6", "6tej":
		variant = "v6"
	case "5", "5t", "5te", "5tej":
		variant = "v5"
	case "4", "4t":
		variant = "v4"
	case "3":
		variant = "v3"
	default:
		variant = "unknown"
	}

	return variant, nil
}

func getAmd64MicroArchLevel() (string, error) {
	flags, err := getCPUInfo("flags")
	if errors.Is(err, errNotFound) {
		return "", fmt.Errorf("failure getting CPU flags: %v", err)
	}

	containsAll := func(set map[string]interface{}, toMatch []string) bool {
		for _, m := range toMatch {
			if _, ok := set[m]; !ok {
				return false
			}
		}
		return true
	}

	flagSet := map[string]interface{}{}
	for _, flag := range strings.Split(flags, " ") {
		flagSet[flag] = true
	}

	// https://unix.stackexchange.com/questions/631217/how-do-i-check-if-my-cpu-supports-x86-64-v2
	level := 1
	if containsAll(flagSet, []string{"lm", "cmov", "cx8", "fpu", "fxsr", "mmx", "syscall", "sse2"}) {
		level = 1
	}
	if level == 1 && containsAll(flagSet, []string{"cx16", "lahf_lm", "popcnt", "sse4_1", "sse4_2", "ssse3"}) {
		level = 2
	}
	if level == 2 && containsAll(flagSet, []string{"avx", "avx2", "bmi1", "bmi2", "f16c", "fma", "abm", "movbe", "xsave"}) {
		level = 3
	}
	if level == 3 && containsAll(flagSet, []string{"avx512f", "avx512bw", "avx512cd", "avx512dq", "avx512vl"}) {
		level = 4
	}
	return fmt.Sprintf("v%d", level), nil
}
