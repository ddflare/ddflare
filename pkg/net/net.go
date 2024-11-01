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

func Resolve(domain string) (string, error) {
	addr, err := net.LookupHost(domain)
	if err != nil {
		return "", fmt.Errorf("cannot resolve %q: %w", domain, err)
	}

	// Returns the first address only
	// TODO: check the number of returned addresses and pass back the IPv4 one
	return addr[0], nil
}
