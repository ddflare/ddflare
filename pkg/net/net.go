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
	"fmt"
	"io"
	"net"
	"net/http"
)

func GetMyPub() (string, error) {
	res, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	ip, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func Resolve(fqdn string) (string, error) {
	addr, err := net.LookupHost(fqdn)
	if err != nil {
		return "", fmt.Errorf("cannot resolve %q: %w", fqdn, err)
	}
	// Return first IPv4 address
	for _, a := range addr {
		if net.ParseIP(a).To4() != nil {
			return a, nil
		}
	}
	return "", fmt.Errorf("no IPv4 address found for %q", fqdn)
}
