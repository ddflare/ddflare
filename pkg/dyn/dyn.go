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

// Package dyndns implements a compliant Dyn updater following the API
// specified at https://help.dyn.com/remote-access-api/ .
package dyn

import (
	"fmt"
	"log/slog"

	"github.com/fgiudici/ddflare/pkg/ddman"
	"github.com/fgiudici/ddflare/pkg/dyndnsapi"
	"github.com/fgiudici/ddflare/pkg/net"
	"github.com/fgiudici/ddflare/pkg/version"
)

var _ ddman.DNSManager = (*Client)(nil)

const (
	defaultAPIEP     = "https://members.dyndns.org"
	defaultUserAgent = "ddflare-dynlib-"
)

type Client struct {
	endpoint  string
	userAgent string
	*dyndnsapi.API
}

// NewWithEndpoint initializes a new Dyn update protocol client which uses 'endpoint' as API
// endpoint.
func NewWithEndpoint(ep string) *Client {
	return &Client{
		endpoint:  ep,
		userAgent: defaultUserAgent + version.Version,
	}
}

func New() *Client {
	return NewWithEndpoint(defaultAPIEP)
}

// GetApiEndpoint returns the current API EndPoint.
func (c *Client) GetApiEndpoint() string {
	return c.endpoint
}

// SetApiEndpoint sets the API Endpoint but would be uneffective if the .Init() has already
// been called on the client.
func (c *Client) SetApiEndpoint(ep string) {
	c.endpoint = ep
}

func (c *Client) GetUserAgent() string {
	return c.userAgent
}

func (c *Client) SetUserAgent(ua string) {
	c.userAgent = ua
}

// Init
func (c *Client) Init(authToken string) error {
	var err error
	if c.API, err = dyndnsapi.New(c.endpoint, authToken, c.userAgent); err != nil {
		return err
	}
	return nil
}

// Resolve returns the current IP address assigned to the FQDN passed as
// parameter.
func (c *Client) Resolve(fqdn string) (string, error) {
	return net.Resolve(fqdn)
}

// Update updates the `fqdn` to the `ip` address passed as parameter.
func (c *Client) Update(fqdn, ip string) error {
	var err error
	var retCode dyndnsapi.ReturnCode
	log := slog.Default().With("endpoint", c.endpoint, "fqdn", fqdn)

	if retCode, err = c.API.Update(fqdn, ip); err != nil {
		return fmt.Errorf("dyn update failed: %w", err)
	}
	if retCode == dyndnsapi.MsgNoChg {
		log.Warn("Dyn API Endpoint replied the FQDN was already set at the right IP", "ip", ip)
	}
	return nil
}
