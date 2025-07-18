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

package dyndnsapi

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		endpoint    string
		token       string
		userAgent   string
		expectError bool
		errorMsg    string
	}{
		"valid_params": {
			endpoint:    "https://api.example.com",
			token:       "user:password",
			userAgent:   "TestApp/1.0",
			expectError: false,
		},
		"empty_endpoint": {
			endpoint:    "",
			token:       "user:password",
			userAgent:   "TestApp/1.0",
			expectError: true,
			errorMsg:    "missing endpoint",
		},
		"empty_useragent": {
			endpoint:    "https://api.example.com",
			token:       "user:password",
			userAgent:   "",
			expectError: true,
			errorMsg:    "missing useragent",
		},
		"empty_token_allowed": {
			endpoint:    "https://api.example.com",
			token:       "",
			userAgent:   "TestApp/1.0",
			expectError: false,
		},
		"all_empty_except_endpoint_and_useragent": {
			endpoint:    "https://api.example.com",
			token:       "",
			userAgent:   "TestApp/1.0",
			expectError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			api, err := New(tt.endpoint, tt.token, tt.userAgent)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
				if api != nil {
					t.Error("Expected API to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if api == nil {
					t.Error("Expected non-nil API")
				} else {
					if api.baseURL != tt.endpoint {
						t.Errorf("Expected baseURL %q, got %q", tt.endpoint, api.baseURL)
					}
					if api.userAgent != tt.userAgent {
						t.Errorf("Expected userAgent %q, got %q", tt.userAgent, api.userAgent)
					}

					// Check that token is properly base64 encoded
					expectedToken := base64.StdEncoding.EncodeToString([]byte(tt.token))
					if api.apiToken != expectedToken {
						t.Errorf("Expected encoded token %q, got %q", expectedToken, api.apiToken)
					}
				}
			}
		})
	}
}

func TestAPI_Update_WithMockServer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		responseBody   string
		responseStatus int
		expectedCode   ReturnCode
		expectError    bool
		errorMsg       string
	}{
		"good_response": {
			responseBody:   "good 192.168.1.1",
			responseStatus: 200,
			expectedCode:   MsgGood,
			expectError:    false,
		},
		"nochg_response": {
			responseBody:   "nochg 192.168.1.1",
			responseStatus: 200,
			expectedCode:   MsgNoChg,
			expectError:    false,
		},
		"badauth_response": {
			responseBody:   "badauth",
			responseStatus: 200,
			expectedCode:   MsgBadAuth,
			expectError:    true,
			errorMsg:       "bad username or password",
		},
		"notdonator_response": {
			responseBody:   "!donator",
			responseStatus: 200,
			expectedCode:   MsgNotDonator,
			expectError:    true,
			errorMsg:       "premium option not available",
		},
		"notfqdn_response": {
			responseBody:   "nofqdn",
			responseStatus: 200,
			expectedCode:   MsgNotFQDN,
			expectError:    true,
			errorMsg:       "invalid FQDN: bad syntax",
		},
		"nohost_response": {
			responseBody:   "nohost",
			responseStatus: 200,
			expectedCode:   MsgNoHost,
			expectError:    true,
			errorMsg:       "invalid FQDN: hostname does not exist",
		},
		"numhost_response": {
			responseBody:   "numhost",
			responseStatus: 200,
			expectedCode:   MsgNumHost,
			expectError:    true,
			errorMsg:       "round robin update detected",
		},
		"abuse_response": {
			responseBody:   "abuse",
			responseStatus: 200,
			expectedCode:   MsgAbuse,
			expectError:    true,
			errorMsg:       "FQDN blocked for update abuse",
		},
		"badagent_response": {
			responseBody:   "badagent",
			responseStatus: 200,
			expectedCode:   MsgBadAgent,
			expectError:    true,
			errorMsg:       "invalid user agent",
		},
		"dnserr_response": {
			responseBody:   "dnserr",
			responseStatus: 200,
			expectedCode:   MsgDNSErr,
			expectError:    true,
			errorMsg:       "server unavailable: DNS error",
		},
		"911_response": {
			responseBody:   "911",
			responseStatus: 200,
			expectedCode:   Msg911,
			expectError:    true,
			errorMsg:       "server unavailable: generic error",
		},
		"unknown_response": {
			responseBody:   "unknown_status",
			responseStatus: 200,
			expectedCode:   MsgUnknownErr,
			expectError:    true,
			errorMsg:       "protocol error: unknown reply message",
		},
		"http_error": {
			responseBody:   "",
			responseStatus: 500,
			expectedCode:   MsgCommErr,
			expectError:    true,
			errorMsg:       "returned 500",
		},
		"empty_response": {
			responseBody:   "",
			responseStatus: 200,
			expectedCode:   MsgCommErr,
			expectError:    true,
			errorMsg:       "failure reading endpoint",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/nic/update" {
					t.Errorf("Expected path /nic/update, got %s", r.URL.Path)
				}

				// Check query parameters
				hostname := r.URL.Query().Get("hostname")
				if hostname != "test.example.com" {
					t.Errorf("Expected hostname test.example.com, got %s", hostname)
				}
				myip := r.URL.Query().Get("myip")
				if myip != "192.168.1.1" {
					t.Errorf("Expected myip 192.168.1.1, got %s", myip)
				}

				// Check headers
				auth := r.Header.Get("Authorization")
				if !strings.HasPrefix(auth, "Basic ") {
					t.Errorf("Expected Basic auth header, got %s", auth)
				}
				ua := r.Header.Get("User-Agent")
				if ua != "TestApp/1.0" {
					t.Errorf("Expected User-Agent TestApp/1.0, got %s", ua)
				}

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			// Create API with mock server endpoint
			api, err := New(server.URL, "user:password", "TestApp/1.0")
			if err != nil {
				t.Fatalf("Failed to create API instance: %v", err)
			}

			code, err := api.Update("test.example.com", "192.168.1.1")

			if code != tt.expectedCode {
				t.Errorf("Expected return code %d, got %d", tt.expectedCode, code)
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestInterpretResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		response     string
		expectedCode ReturnCode
	}{
		"good":            {"good", MsgGood},
		"good_with_ip":    {"good 192.168.1.1", MsgGood},
		"nochg":           {"nochg", MsgNoChg},
		"nochg_with_ip":   {"nochg 192.168.1.1", MsgNoChg},
		"badauth":         {"badauth", MsgBadAuth},
		"notdonator":      {"!donator", MsgNotDonator},
		"notfqdn":         {"nofqdn", MsgNotFQDN},
		"nohost":          {"nohost", MsgNoHost},
		"numhost":         {"numhost", MsgNumHost},
		"abuse":           {"abuse", MsgAbuse},
		"badagent":        {"badagent", MsgBadAgent},
		"dnserr":          {"dnserr", MsgDNSErr},
		"911":             {"911", Msg911},
		"unknown":         {"unknown_status", MsgUnknownErr},
		"empty":           {"", MsgUnknownErr},
		"multiple_fields": {"good 192.168.1.1 extra_field", MsgGood},
		"whitespace":      {"  good  192.168.1.1  ", MsgGood},
		"case_sensitive":  {"Good", MsgUnknownErr}, // Should be case sensitive
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			code := interpretResponse(tt.response)

			if code != tt.expectedCode {
				t.Errorf("Expected return code %d for response %q, got %d", tt.expectedCode, tt.response, code)
			}
		})
	}
}

func TestReturnCodeConstants(t *testing.T) {
	t.Parallel()

	// Test that all return codes have corresponding messages
	expectedCodes := []ReturnCode{
		MsgGood, MsgNoChg, MsgBadAuth, MsgNotDonator, MsgNotFQDN,
		MsgNoHost, MsgNumHost, MsgAbuse, MsgBadAgent, MsgDNSErr,
		Msg911, MsgUnknownErr,
	}

	for _, code := range expectedCodes {
		if msg, exists := code2Msg[int(code)]; !exists {
			t.Errorf("Return code %d does not have a corresponding message", code)
		} else if msg == "" {
			t.Errorf("Return code %d has an empty message", code)
		}
	}

	// Test that MsgLast is greater than all other codes
	for _, code := range expectedCodes {
		if code >= MsgLast {
			t.Errorf("Return code %d should be less than MsgLast (%d)", code, MsgLast)
		}
	}
}

func TestCode2MsgMapping(t *testing.T) {
	t.Parallel()

	expectedMappings := map[int]string{
		MsgGood:       "the update was successful",
		MsgNoChg:      "the update changed no settings",
		MsgBadAuth:    "bad username or password",
		MsgNotDonator: "premium option not available for this account",
		MsgNotFQDN:    "invalid FQDN: bad syntax",
		MsgNoHost:     "invalid FQDN: hostname does not exist",
		MsgNumHost:    "round robin update detected",
		MsgAbuse:      "FQDN blocked for update abuse",
		MsgBadAgent:   "invalid user agent",
		MsgDNSErr:     "server unavailable: DNS error",
		Msg911:        "server unavailable: generic error",
		MsgUnknownErr: "protocol error: unknown reply message",
	}

	for code, expectedMsg := range expectedMappings {
		if msg, exists := code2Msg[code]; !exists {
			t.Errorf("Expected message for code %d not found", code)
		} else if msg != expectedMsg {
			t.Errorf("Expected message %q for code %d, got %q", expectedMsg, code, msg)
		}
	}
}

func TestAPI_Update_AuthHeaderEncoding(t *testing.T) {
	t.Parallel()

	// Test that the authorization header is properly base64 encoded
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			t.Errorf("Expected Basic auth, got %s", auth)
			return
		}

		// Decode the base64 part
		encoded := strings.TrimPrefix(auth, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Errorf("Failed to decode auth header: %v", err)
			return
		}

		if string(decoded) != "testuser:testpass" {
			t.Errorf("Expected decoded auth 'testuser:testpass', got %q", string(decoded))
		}

		w.WriteHeader(200)
		w.Write([]byte("good 192.168.1.1"))
	}))
	defer server.Close()

	api, err := New(server.URL, "testuser:testpass", "TestApp/1.0")
	if err != nil {
		t.Fatalf("Failed to create API instance: %v", err)
	}

	_, err = api.Update("test.example.com", "192.168.1.1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAPI_Update_ConnectionError(t *testing.T) {
	t.Parallel()

	// Test with invalid endpoint that should cause connection error
	api, err := New("http://invalid-endpoint-that-should-not-exist.local:9999", "user:pass", "TestApp/1.0")
	if err != nil {
		t.Fatalf("Failed to create API instance: %v", err)
	}

	code, err := api.Update("test.example.com", "192.168.1.1")

	if code != MsgCommErr {
		t.Errorf("Expected return code %d, got %d", MsgCommErr, code)
	}
	if err == nil {
		t.Error("Expected connection error but got none")
	}
	if !strings.Contains(err.Error(), "connection to") {
		t.Errorf("Expected connection error message, got %q", err.Error())
	}
}
