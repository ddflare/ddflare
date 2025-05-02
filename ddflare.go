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

// Package ddflare exposes functions to manage DDNS (Dynamic DNS) updates.
package ddflare

import (
	"fmt"

	"github.com/ddflare/ddflare/pkg/cflare"
	"github.com/ddflare/ddflare/pkg/ddman"
	"github.com/ddflare/ddflare/pkg/dyn"
	"github.com/ddflare/ddflare/pkg/net"
)

// DNSManagerType identifies the service type used for DDNS updates.
type DNSManagerType int

const (
	Cloudflare = iota
	Dyn
	DDNS
	NoIP
)

// DNSManager represents a DDNS service instance and exposes the methods
// to read and update the managed DNS records.
type DNSManager struct {
	ddman.DNSManager
	lastSetAddresses map[string]string
}

// GetPublicIP returns the current Public IP address by querying the
// "api.ipify.org" service.
func GetPublicIP() (string, error) {
	var (
		ip  string
		err error
	)

	if ip, err = net.GetMyPub(); err != nil {
		return "", fmt.Errorf("cannot retrieve public address: %w", err)
	}

	return ip, nil
}

// Resolve returns the IP address of the FQDN passed as argument using the
// local resolver.
func Resolve(fqdn string) (string, error) {
	return net.Resolve(fqdn)
}

// NewDNSManager() returns a new DNSManager of the give DNSManagerType.
// It returns an error which is not nil only if a wrong DNSManagerType
// is passed to NewDNSManager.
func NewDNSManager(dt DNSManagerType) (*DNSManager, error) {
	dm := &DNSManager{}

	switch dt {
	case Cloudflare:
		dm.DNSManager = cflare.New()
	case Dyn:
		dm.DNSManager = dyn.New()
	case DDNS:
		dm.DNSManager = dyn.NewWithEndpoint("https://update.ddns.org")
	case NoIP:
		dm.DNSManager = dyn.NewWithEndpoint("https://dynupdate.no-ip.com")
	default:
		return nil, fmt.Errorf("invalid DNS manager backend (%d)", dt)
	}

	dm.lastSetAddresses = make(map[string]string)
	return dm, nil
}

// UpdateFQDN() updates `fqdn` to `ip` using the DNSManager backend.
// The `fqdn` and `ip` address are stored in a local cache so that
// the update operation can be skipped if the `fqdn` and `ip` addresses
// are the same of the previous operation.
func (d *DNSManager) UpdateFQDN(fqdn, ip string) error {
	if ip == d.lastSetAddresses[fqdn] {
		return nil
	}
	if err := d.Update(fqdn, ip); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	d.lastSetAddresses[fqdn] = ip
	return nil
}

// IsFQDNUpToDate() checks if the `fqdn` was already set to the desired `ip`.
// First the local cache is checked for previously updated value: if local cache
// is different, then the `fqdn` is resolved and checked against the passed `ip`.
func (d *DNSManager) IsFQDNUpToDate(fqdn, ip string) (bool, error) {
	var (
		resIP string
		err   error
	)
	if ip == d.lastSetAddresses[fqdn] {
		return true, nil
	}
	if resIP, err = d.Resolve(fqdn); err != nil {
		return false, fmt.Errorf("resolve failed: %w", err)
	}
	if resIP == ip {
		return true, nil
	}

	return false, nil
}
