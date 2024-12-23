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

	// Returns the first address only
	// TODO: check the number of returned addresses and pass back the IPv4 one
	return addr[0], nil
}
