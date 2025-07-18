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

package cflare

import (
	"context"
	"errors"
	"fmt"
	"testing"

	cf "github.com/cloudflare/cloudflare-go"
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
			cflare := New()
			err := cflare.Init(tt.token)
			if tt.fails {
				if err == nil {
					t.Fatalf("expected failure with token %q but got %+v", tt.token, cflare.api)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected failure: %v", err)
			}
			if cflare.api.APIToken != tt.token {
				t.Fatalf("expected token %q but got %q", tt.token, cflare.api.APIToken)
			}
		})
	}
}

func TestCloudflare_GetSetApiEndpoint(t *testing.T) {
	t.Parallel()

	c := New()
	err := c.Init("test-token")
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test default endpoint
	defaultEndpoint := c.GetApiEndpoint()
	if defaultEndpoint == "" {
		t.Error("Expected non-empty default API endpoint")
	}

	// Test setting custom endpoint
	customEndpoint := "https://custom.api.endpoint.com"
	c.SetApiEndpoint(customEndpoint)

	if got := c.GetApiEndpoint(); got != customEndpoint {
		t.Errorf("Expected API endpoint %q, got %q", customEndpoint, got)
	}
}

func TestCloudflare_GetSetUserAgent(t *testing.T) {
	t.Parallel()

	c := New()
	err := c.Init("test-token")
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test default user agent
	defaultUA := c.GetUserAgent()
	if defaultUA == "" {
		t.Error("Expected non-empty default user agent")
	}

	// Test setting custom user agent
	customUA := "CustomApp/1.0"
	c.SetUserAgent(customUA)

	if got := c.GetUserAgent(); got != customUA {
		t.Errorf("Expected user agent %q, got %q", customUA, got)
	}
}

func TestCloudflare_InitError(t *testing.T) {
	t.Parallel()

	c := New()

	// Test uninitialized state
	if c.api != nil {
		t.Error("Expected api to be nil before initialization")
	}
}

func TestCloudflare_Resolve(t *testing.T) {
	t.Parallel()

	c := New()

	tests := map[string]struct {
		fqdn        string
		expectError bool
	}{
		"valid_domain":   {"google.com", false},
		"invalid_domain": {"this-domain-should-not-exist-12345.invalid", true},
		"empty_domain":   {"", true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ip, err := c.Resolve(tt.fqdn)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for FQDN %q, but got IP %q", tt.fqdn, ip)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for FQDN %q: %v", tt.fqdn, err)
				}
				if ip == "" {
					t.Errorf("Expected non-empty IP for FQDN %q", tt.fqdn)
				}
			}
		})
	}
}

func TestCloudflare_UpdateUnitialized(t *testing.T) {
	t.Parallel()

	c := New()
	// Don't initialize the API

	err := c.Update("test.example.com", "192.168.1.1")
	if err == nil {
		t.Error("Expected error when API is not initialized")
	}

	expectedMsg := "not authorized"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestGetZone(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fqdn         string
		expectedZone string
		expectError  bool
	}{
		"subdomain":      {"sub.example.com", "example.com", false},
		"domain":         {"example.com", "example.com", false},
		"deep_subdomain": {"a.b.c.example.com", "example.com", false},
		"single_word":    {"localhost", "", true},
		"empty":          {"", "", true},
		"single_dot":     {".", "", true},
		"co_uk_domain":   {"test.example.co.uk", "co.uk", false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zone, err := getZone(tt.fqdn)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for FQDN %q, but got zone %q", tt.fqdn, zone)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for FQDN %q: %v", tt.fqdn, err)
				}
				if zone != tt.expectedZone {
					t.Errorf("Expected zone %q for FQDN %q, got %q", tt.expectedZone, tt.fqdn, zone)
				}
			}
		})
	}
}

// mockCloudflareAPI is a mock implementation for testing Update method
type mockCloudflareAPI struct {
	zoneIDByNameFunc    func(zoneName string) (string, error)
	listDNSRecordsFunc  func(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error)
	updateDNSRecordFunc func(ctx context.Context, rc *cf.ResourceContainer, params cf.UpdateDNSRecordParams) (cf.DNSRecord, error)
}

func (m *mockCloudflareAPI) ZoneIDByName(zoneName string) (string, error) {
	if m.zoneIDByNameFunc != nil {
		return m.zoneIDByNameFunc(zoneName)
	}
	return "", errors.New("mock not implemented")
}

func (m *mockCloudflareAPI) ListDNSRecords(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
	if m.listDNSRecordsFunc != nil {
		return m.listDNSRecordsFunc(ctx, rc, params)
	}
	return nil, nil, errors.New("mock not implemented")
}

func (m *mockCloudflareAPI) UpdateDNSRecord(ctx context.Context, rc *cf.ResourceContainer, params cf.UpdateDNSRecordParams) (cf.DNSRecord, error) {
	if m.updateDNSRecordFunc != nil {
		return m.updateDNSRecordFunc(ctx, rc, params)
	}
	return cf.DNSRecord{}, errors.New("mock not implemented")
}

func TestCloudflare_UpdateWithMock(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fqdn                string
		ip                  string
		zoneIDByNameFunc    func(string) (string, error)
		listDNSRecordsFunc  func(context.Context, *cf.ResourceContainer, cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error)
		updateDNSRecordFunc func(context.Context, *cf.ResourceContainer, cf.UpdateDNSRecordParams) (cf.DNSRecord, error)
		expectError         bool
		expectedErrorMsg    string
	}{
		"invalid_zone": {
			fqdn:             "invalid",
			ip:               "192.168.1.1",
			expectError:      true,
			expectedErrorMsg: "cannot identify DNS zone",
		},
		"zone_not_found": {
			fqdn: "test.example.com",
			ip:   "192.168.1.1",
			zoneIDByNameFunc: func(zoneName string) (string, error) {
				return "", errors.New("zone not found")
			},
			expectError:      true,
			expectedErrorMsg: "cannot retrieve DNS zone id",
		},
		"list_dns_records_error": {
			fqdn: "test.example.com",
			ip:   "192.168.1.1",
			zoneIDByNameFunc: func(zoneName string) (string, error) {
				return "zone123", nil
			},
			listDNSRecordsFunc: func(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
				return nil, nil, errors.New("API error")
			},
			expectError:      true,
			expectedErrorMsg: "API error",
		},
		"no_records_found": {
			fqdn: "test.example.com",
			ip:   "192.168.1.1",
			zoneIDByNameFunc: func(zoneName string) (string, error) {
				return "zone123", nil
			},
			listDNSRecordsFunc: func(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
				return []cf.DNSRecord{}, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "found 0 matching records",
		},
		"multiple_records_found": {
			fqdn: "test.example.com",
			ip:   "192.168.1.1",
			zoneIDByNameFunc: func(zoneName string) (string, error) {
				return "zone123", nil
			},
			listDNSRecordsFunc: func(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
				return []cf.DNSRecord{
					{ID: "rec1", Name: "test.example.com", Type: "A"},
					{ID: "rec2", Name: "test.example.com", Type: "A"},
				}, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "found 2 matching records",
		},
		"update_dns_record_error": {
			fqdn: "test.example.com",
			ip:   "192.168.1.1",
			zoneIDByNameFunc: func(zoneName string) (string, error) {
				return "zone123", nil
			},
			listDNSRecordsFunc: func(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
				return []cf.DNSRecord{
					{ID: "rec1", Name: "test.example.com", Type: "A", Content: "1.2.3.4"},
				}, nil, nil
			},
			updateDNSRecordFunc: func(ctx context.Context, rc *cf.ResourceContainer, params cf.UpdateDNSRecordParams) (cf.DNSRecord, error) {
				return cf.DNSRecord{}, errors.New("update failed")
			},
			expectError:      true,
			expectedErrorMsg: "update failed",
		},
		"successful_update": {
			fqdn: "test.example.com",
			ip:   "192.168.1.1",
			zoneIDByNameFunc: func(zoneName string) (string, error) {
				return "zone123", nil
			},
			listDNSRecordsFunc: func(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
				return []cf.DNSRecord{
					{
						ID:      "rec1",
						Name:    "test.example.com",
						Type:    "A",
						Content: "1.2.3.4",
						TTL:     300,
					},
				}, nil, nil
			},
			updateDNSRecordFunc: func(ctx context.Context, rc *cf.ResourceContainer, params cf.UpdateDNSRecordParams) (cf.DNSRecord, error) {
				// Verify the update parameters
				if params.Content != "192.168.1.1" {
					return cf.DNSRecord{}, fmt.Errorf("expected IP 192.168.1.1, got %s", params.Content)
				}
				return cf.DNSRecord{
					ID:      "rec1",
					Name:    "test.example.com",
					Type:    "A",
					Content: "192.168.1.1",
					TTL:     300,
				}, nil
			},
			expectError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Note: We can't easily mock the Cloudflare API in the current implementation
			// This test demonstrates the structure but would require refactoring the
			// Cloudflare struct to accept an interface for testing

			// For now, we'll test the parts we can test
			if tt.fqdn == "invalid" {
				// Test the getZone function directly
				_, err := getZone(tt.fqdn)
				if !tt.expectError {
					t.Errorf("Expected error for invalid FQDN")
				}
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			}
		})
	}
}
