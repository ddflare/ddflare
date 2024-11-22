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
	"log"
	"log/slog"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/fgiudici/ddflare/pkg/ddns"
)

var _ ddns.Recorder = (*Cloudflare)(nil)

type Cloudflare struct {
	api *cf.API
}

func New() *Cloudflare {
	return &Cloudflare{}
}

func (c *Cloudflare) Init(token string) error {
	var err error
	// Never returns error when no options are passed (like in this case)
	c.api, err = cf.NewWithAPIToken(token)
	return err
}

func (c *Cloudflare) Write(record, zone, ip string) error {
	if c.api == nil {
		return fmt.Errorf("not authorized")
	}

	ctx := context.Background()

	zoneID, err := c.api.ZoneIDByName(zone)
	if err != nil {
		return err
	}
	log.Printf("Zone ID: %s", zoneID)
	dnsRecs, _, err := c.api.ListDNSRecords(ctx, cf.ZoneIdentifier(zoneID),
		cf.ListDNSRecordsParams{Name: record})
	if err != nil {
		return err
	}
	for i, d := range dnsRecs {
		slog.Debug("record found", "id", i, "data", d)
	}
	if len(dnsRecs) > 1 {
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
	slog.Debug("record updated", "data", rec)

	return nil
}
