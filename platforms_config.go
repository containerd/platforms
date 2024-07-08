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
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

const platformConfigPath = "/etc/containerd/platform-config.toml"

var config *platformConfig

type platformConfig struct {
	Features        []string          `json:"features,omitempty"`
	Compatibilities map[string]string `json:"compatibilities,omitempty"`
}

func readConfig() (*platformConfig, error) {
	if config == nil {
		b, err := os.ReadFile(platformConfigPath)
		if err != nil {
			return nil, err
		}
		if err := toml.Unmarshal(b, config); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func mustReadConfig() *platformConfig {
	p, err := readConfig()
	if err != nil {
		panic(err)
	}
	return p
}
