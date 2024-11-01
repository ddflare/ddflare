/*
Copyright © 2024 Francesco Giudici <dev@foggy.day>

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

package ddns

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		token string
		fails bool
	}{
		"empty": {"", true},
		"short": {"xyz", false},
		"long":  {"13212312312312sdfdsfsdfdsfdsfdsfsdfs123123123123123123", false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cf := Cloudflare{}
			err := cf.New(tt.token)
			if tt.fails {
				if err == nil {
					t.Fatalf("expected failure with token %q but got %+v", tt.token, cf.api)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected failure: %v", err)
			}
			if cf.api.APIToken != tt.token {
				t.Fatalf("expected token %q but got %q", tt.token, cf.api.APIToken)
			}
		})
	}
}
