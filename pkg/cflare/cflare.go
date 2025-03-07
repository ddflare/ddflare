/*
Copyright © 2024 Francesco Giudici <francesco.giudici@suse.com>

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

package cflare

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/fgiudici/ddflare/pkg/ddman"
	"github.com/fgiudici/ddflare/pkg/net"
)

var _ ddman.DNSManager = (*Cloudflare)(nil)

type Cloudflare struct {
	api *cf.API
}

func New() *Cloudflare {
	return &Cloudflare{}
}

func (c *Cloudflare) GetApiEndpoint() string {
	return c.api.BaseURL
}

func (c *Cloudflare) SetApiEndpoint(ep string) {
	c.api.BaseURL = ep
}

func (c *Cloudflare) GetUserAgent() string {
	return c.api.UserAgent
}

func (c *Cloudflare) SetUserAgent(ua string) {
	c.api.UserAgent = ua
}

func (c *Cloudflare) Init(token string) error {
	var err error
	// Never returns error when no options are passed (like in this case)
	c.api, err = cf.NewWithAPIToken(token)
	return err
}

func (c *Cloudflare) Resolve(fqdn string) (string, error) {
	return net.Resolve(fqdn)
}

func (c *Cloudflare) Update(fqdn, ip string) error {
	if c.api == nil {
		return fmt.Errorf("not authorized")
	}

	var err error
	ctx := context.Background()
	log := slog.Default().With("fqdn", fqdn)
	zone := ""

	if zone, err = getZone(fqdn); err != nil {
		return fmt.Errorf("cannot identify DNS zone: %w", err)
	}

	zoneID, err := c.api.ZoneIDByName(zone)
	if err != nil {
		return fmt.Errorf("cannot retrieve DNS zone id: %w", err)
	}

	log.Debug("DNS zone found", "zone", zone, "zoneID", zoneID)

	dnsRecs, _, err := c.api.ListDNSRecords(ctx, cf.ZoneIdentifier(zoneID),
		cf.ListDNSRecordsParams{Name: fqdn})
	if err != nil {
		return err
	}
	for _, d := range dnsRecs {
		log.Debug("record found", "data", d)
	}
	if len(dnsRecs) != 1 {
		return fmt.Errorf("found %d matching records", len(dnsRecs))
	}
	rec := dnsRecs[0]

	updateRec := cf.UpdateDNSRecordParams{
		Type:    rec.Type,
		Name:    rec.Name,
		Content: ip,
		Data:    rec.Data,
		ID:      rec.ID,
		Tags:    rec.Tags,
		TTL:     rec.TTL,
	}

	if rec, err = c.api.UpdateDNSRecord(ctx, cf.ZoneIdentifier(zoneID), updateRec); err != nil {
		return err
	}
	log.Debug("record updated", "data", rec)

	return nil
}

func getZone(fqdn string) (string, error) {
	domain := strings.Split(fqdn, ".")
	if len(domain) < 2 {
		return "", fmt.Errorf("%q is not a valid dns name", fqdn)
	}
	zone := domain[len(domain)-2] + "." + domain[len(domain)-1]
	return zone, nil
}
