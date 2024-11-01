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

package net

import (
	"io"
	"net/http"
	"testing"
)

func TestGetMyPub(t *testing.T) {
	t.Parallel()
	var myIp []byte

	res, err := http.Get("https://api.ipify.org")
	if err == nil {
		myIp, err = io.ReadAll(res.Body)
	}
	if err != nil {
		t.Skipf("Skipping test \"TestGetMyPub\" as cannot retrieve pub ip: %v", err)
	}

	ip, err := GetMyPub()
	if err != nil {
		t.Fatal(err)
	}
	if ip != string(myIp) {
		t.Fatalf("expecting %q, got %q", string(myIp), ip)
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		domain string
		ip     string
		fails  bool
	}{
		"sslip.io": {
			"10.10.10.10.sslip.io",
			"10.10.10.10",
			false,
		},
		"nonexistent": {
			"notexistent.domain",
			"",
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			res, err := Resolve(test.domain)
			t.Logf("Resolve(%q): %q, error: %v", test.domain, res, err)
			if test.fails {
				if err == nil {
					t.Fatalf("Resolve(%q): expecting error, got %q", test.domain, res)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", test.domain, err)
			}

			if test.ip != res {
				t.Fatalf("Resolve(%q): expecting %q, got %q", test.domain, res, test.ip)
			}
		})
	}
}
