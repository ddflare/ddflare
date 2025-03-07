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

// Package dyndnsapi implements the basic Dynamic DNS API from DynDNS specified at
// https://help.dyn.com/remote-access-api/ .
package dyndnsapi

import (
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type API struct {
	apiToken  string // base64 encoded
	baseURL   string
	userAgent string
}

type ReturnCode int

// Msg consts track the DynDNS API return codes: https://help.dyn.com/remote-access-api/return-codes/
const (
	MsgGood       = iota // The update was successful, and the hostname is now updated.
	MsgNoChg             // The update changed no settings, and is considered abusive. Additional nochg updates will cause the hostname to become blocked.
	MsgBadAuth           // The username and password pair do not match a real user.
	MsgNotDonator        // An option available only to credited users (such as offline URL) was specified, but the user is not a credited user.
	MsgNotFQDN           // The hostname specified is not a fully-qualified domain name (not in the form hostname.dyndns.org or domain.com).
	MsgNoHost            // The hostname specified does not exist in this user account (or is not in the service specified in the system parameter).
	MsgNumHost           // Too many hosts (more than 20) specified in an update. Also returned if trying to update a round robin (which is not allowed).
	MsgAbuse             // The hostname specified is blocked for update abuse.
	MsgBadAgent          // The user agent was not sent or HTTP method is not permitted (we recommend use of GET request method).
	MsgDNSErr            // DNS error encountered. Retry not before 30 min.
	Msg911               // There is a problem or scheduled maintenance on our side. Retry not before 30 min.
	MsgUnknownErr        // Unknown message received.
	MsgDataErr           // This tracks missing data, is no a return code.
	MsgCommErr           // This tracks dial-in issues, is not a return code.
	MsgLast              // Keep always as the last!
)

var code2Msg = map[int]string{
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

// New initializes a new DynDNS update connection to the speficied 'endpoint', authenticating
// with the token parameter and identifying the client via the 'useragent' string.
func New(endpoint, token, useragent string) (*API, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("cannot instantiate new dyndns API: missing endpoint")
	}
	if useragent == "" {
		return nil, fmt.Errorf("cannot instantiate new dyndns API: missing useragent")
	}
	return &API{
		baseURL:   endpoint,
		apiToken:  base64.StdEncoding.EncodeToString([]byte(token)),
		userAgent: useragent,
	}, nil
}

// Update updates the `fqdn` to the `ip` address passed as parameters.
func (c *API) Update(fqdn, ip string) (ReturnCode, error) {
	if c.apiToken == "" {
		return MsgDataErr, fmt.Errorf("no authorization credentials found")
	}
	if fqdn == "" {
		return MsgDataErr, fmt.Errorf("fqdn is missing")
	}
	if ip == "" {
		return MsgDataErr, fmt.Errorf("ip address is missing")
	}

	var (
		req *http.Request
		res *http.Response
		err error
	)

	log := slog.Default().With("endpoint", c.baseURL, "fqdn", fqdn)

	if req, err = http.NewRequest("GET", c.baseURL+"/nic/update", nil); err != nil {
		return MsgCommErr, fmt.Errorf("connection to %s failed: %w", c.baseURL, err)
	}
	req.Header.Add("Authorization", "Basic "+c.apiToken)
	req.Header.Add("User-Agent", c.userAgent)

	q := req.URL.Query()
	q.Add("hostname", fqdn)
	q.Add("myip", ip)

	req.URL.RawQuery = q.Encode()
	if res, err = http.DefaultClient.Do(req); err != nil {
		return MsgCommErr, fmt.Errorf("connection to %s failed: %w", c.baseURL, err)
	}
	defer res.Body.Close()

	log.Debug("endpoint connected", "status", res.Status, "code", res.StatusCode)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return MsgCommErr, fmt.Errorf("endpoint %q returned %d (%s) status", c.baseURL, res.StatusCode, res.Status)
	}

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil || len(body) == 0 {
		return MsgCommErr, fmt.Errorf("failure reading endpoint %q reply: %w", c.baseURL, err)
	}
	log.Debug("parsing reply message", "body", string(body))

	retCode := interpretResponse(string(body))

	if retCode == MsgGood || retCode == MsgNoChg {
		return retCode, nil
	}
	return retCode, fmt.Errorf("%s", code2Msg[int(retCode)])
}

// interpretResponse decodes the returned status messages and reports back to the caller
// the return code received.
func interpretResponse(resp string) ReturnCode {
	// let's ensure we don't get an empty string... we check in the caller, but better
	// stay safe for any future change may happen in the code.
	if resp == "" {
		return MsgUnknownErr
	}

	respSlice := strings.Fields(resp)
	respLen := len(respSlice)
	if respLen > 2 {
		slog.Warn("unexpected number of arguments in reply", "reply", respSlice)
	}
	status := respSlice[0]

	switch status {
	case "good":
		return MsgGood
	case "nochg":
		return MsgNoChg
	case "badauth":
		return MsgBadAuth
	case "!donator":
		return MsgNotDonator
	case "nofqdn":
		return MsgNotFQDN
	case "nohost":
		return MsgNoHost
	case "numhost":
		return MsgNumHost
	case "abuse":
		return MsgAbuse
	case "badagent":
		return MsgBadAgent
	case "dnserr":
		return MsgDNSErr
	case "911":
		return Msg911
	default:
		return MsgUnknownErr
	}
}
