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

package dyn

import (
	"strings"
	"testing"

	"github.com/ddflare/ddflare/pkg/ddman"
	"github.com/ddflare/ddflare/pkg/version"
)

func TestNew(t *testing.T) {
	t.Parallel()

	client := New()

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	expectedEndpoint := "https://members.dyndns.org"
	if client.GetApiEndpoint() != expectedEndpoint {
		t.Errorf("Expected default endpoint %q, got %q", expectedEndpoint, client.GetApiEndpoint())
	}

	expectedUserAgent := "ddflare-dynlib-" + version.Version
	if client.GetUserAgent() != expectedUserAgent {
		t.Errorf("Expected default user agent %q, got %q", expectedUserAgent, client.GetUserAgent())
	}

	if client.API != nil {
		t.Error("Expected API to be nil before initialization")
	}
}

func TestNewWithEndpoint(t *testing.T) {
	t.Parallel()

	customEndpoint := "https://custom.dyn.endpoint.com"
	client := NewWithEndpoint(customEndpoint)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.GetApiEndpoint() != customEndpoint {
		t.Errorf("Expected custom endpoint %q, got %q", customEndpoint, client.GetApiEndpoint())
	}

	expectedUserAgent := "ddflare-dynlib-" + version.Version
	if client.GetUserAgent() != expectedUserAgent {
		t.Errorf("Expected default user agent %q, got %q", expectedUserAgent, client.GetUserAgent())
	}

	if client.API != nil {
		t.Error("Expected API to be nil before initialization")
	}
}

func TestClient_ImplementsDNSManager(t *testing.T) {
	t.Parallel()

	var _ ddman.DNSManager = (*Client)(nil)
}

func TestClient_GetSetApiEndpoint(t *testing.T) {
	t.Parallel()

	client := New()

	// Test default endpoint
	defaultEndpoint := client.GetApiEndpoint()
	if defaultEndpoint == "" {
		t.Error("Expected non-empty default API endpoint")
	}

	// Test setting custom endpoint
	customEndpoint := "https://custom.api.endpoint.com"
	client.SetApiEndpoint(customEndpoint)

	if got := client.GetApiEndpoint(); got != customEndpoint {
		t.Errorf("Expected API endpoint %q, got %q", customEndpoint, got)
	}
}

func TestClient_GetSetUserAgent(t *testing.T) {
	t.Parallel()

	client := New()

	// Test default user agent
	defaultUA := client.GetUserAgent()
	if defaultUA == "" {
		t.Error("Expected non-empty default user agent")
	}
	if !strings.HasPrefix(defaultUA, "ddflare-dynlib-") {
		t.Errorf("Expected user agent to start with 'ddflare-dynlib-', got %q", defaultUA)
	}

	// Test setting custom user agent
	customUA := "CustomApp/1.0"
	client.SetUserAgent(customUA)

	if got := client.GetUserAgent(); got != customUA {
		t.Errorf("Expected user agent %q, got %q", customUA, got)
	}
}

func TestClient_Init(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		authToken   string
		expectError bool
	}{
		"valid_token":   {"valid-token-123", false},
		"empty_token":   {"", false},
		"another_token": {"user:password", false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := New()
			err := client.Init(tt.authToken)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if client.API != nil {
					t.Error("Expected API to remain nil after failed initialization")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client.API == nil {
					t.Error("Expected API to be initialized")
				}
			}
		})
	}
}

func TestClient_InitAfterSetEndpoint(t *testing.T) {
	t.Parallel()

	customEndpoint := "https://custom.dyn.endpoint.com"
	client := New()
	client.SetApiEndpoint(customEndpoint)

	err := client.Init("test-token")
	if err != nil {
		t.Fatalf("Unexpected error during initialization: %v", err)
	}

	if client.API == nil {
		t.Error("Expected API to be initialized")
	}

	// Verify that the endpoint was used during initialization
	if client.GetApiEndpoint() != customEndpoint {
		t.Errorf("Expected endpoint %q to be preserved, got %q", customEndpoint, client.GetApiEndpoint())
	}
}

func TestClient_InitAfterSetUserAgent(t *testing.T) {
	t.Parallel()

	customUA := "TestApp/2.0"
	client := New()
	client.SetUserAgent(customUA)

	err := client.Init("test-token")
	if err != nil {
		t.Fatalf("Unexpected error during initialization: %v", err)
	}

	if client.API == nil {
		t.Error("Expected API to be initialized")
	}

	// Verify that the user agent was preserved
	if client.GetUserAgent() != customUA {
		t.Errorf("Expected user agent %q to be preserved, got %q", customUA, client.GetUserAgent())
	}
}

func TestClient_Resolve(t *testing.T) {
	t.Parallel()

	client := New()

	tests := map[string]struct {
		fqdn        string
		expectError bool
	}{
		"valid_domain":   {"google.com", false},
		"localhost":      {"localhost", false},
		"invalid_domain": {"this-domain-should-not-exist-12345.invalid", true},
		"empty_domain":   {"", true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ip, err := client.Resolve(tt.fqdn)

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

func TestClient_SetEndpointAfterInit(t *testing.T) {
	t.Parallel()

	client := New()

	// Initialize the client
	err := client.Init("test-token")
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// Set a new endpoint after initialization
	newEndpoint := "https://new.endpoint.com"
	client.SetApiEndpoint(newEndpoint)

	// Verify the endpoint was changed
	if client.GetApiEndpoint() != newEndpoint {
		t.Errorf("Expected endpoint %q, got %q", newEndpoint, client.GetApiEndpoint())
	}

	// Note: According to the comment in the code, this change would be
	// ineffective for the already initialized API, but the getter should
	// still return the new value
}

func TestClient_FullWorkflow(t *testing.T) {
	t.Parallel()

	// Test the complete workflow without actual API calls
	client := New()

	// 1. Configure client
	customEndpoint := "https://localhost:8080"
	customUA := "TestWorkflow/1.0"

	client.SetApiEndpoint(customEndpoint)
	client.SetUserAgent(customUA)

	// Verify configuration
	if client.GetApiEndpoint() != customEndpoint {
		t.Errorf("Expected endpoint %q, got %q", customEndpoint, client.GetApiEndpoint())
	}
	if client.GetUserAgent() != customUA {
		t.Errorf("Expected user agent %q, got %q", customUA, client.GetUserAgent())
	}

	// 2. Initialize client
	err := client.Init("test-token")
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// 3. Test resolve (this should work with real domains)
	ip, err := client.Resolve("google.com")
	if err != nil {
		t.Errorf("Failed to resolve google.com: %v", err)
	} else if ip == "" {
		t.Error("Expected non-empty IP for google.com")
	}

	// 4. Test update (this will likely fail with test credentials, but tests the flow)
	err = client.Update("test.example.com", "192.168.1.1")
	// We don't assert on the error here since it depends on external API
	// But we can verify the method doesn't panic
	t.Logf("Update result: %v", err)
}

func TestDefaultConstants(t *testing.T) {
	t.Parallel()

	// Test that the default constants are as expected
	client := New()

	expectedEndpoint := "https://members.dyndns.org"
	if client.GetApiEndpoint() != expectedEndpoint {
		t.Errorf("Expected default endpoint %q, got %q", expectedEndpoint, client.GetApiEndpoint())
	}

	expectedUAPrefix := "ddflare-dynlib-"
	ua := client.GetUserAgent()
	if !strings.HasPrefix(ua, expectedUAPrefix) {
		t.Errorf("Expected user agent to start with %q, got %q", expectedUAPrefix, ua)
	}

	if !strings.HasSuffix(ua, version.Version) {
		t.Errorf("Expected user agent to end with version %q, got %q", version.Version, ua)
	}
}
