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

package dyndns

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fgiudici/ddflare/pkg/ddman"
	"github.com/fgiudici/ddflare/pkg/net"
	"github.com/fgiudici/ddflare/pkg/version"
)

var _ ddman.DNSManager = (*DynDNS)(nil)

type DynDNS struct {
	endpoint  string
	authToken string // base64 encoded
}

// New initializes a new DynDNS update protocol backend which forwards update requests
// to the endpoint parameter passed as argument.
func New(endpoint string) *DynDNS {
	return &DynDNS{endpoint: endpoint}
}

func (d *DynDNS) Add(fqdn string) error {
	return fmt.Errorf("not supported")
}

func (d *DynDNS) Del(fqdn string) error {
	return fmt.Errorf("not supported")
}

func (d *DynDNS) GetApiEndpoint() string {
	return d.endpoint
}

// Init expects a '$username:$password' token which once encoded base64
// could be used as authentication token for the DynDNS update protocol
// (to be passed in the Authorization Header of the HTTP GET request).
// It cannot fail, so it always returns nil.
func (d *DynDNS) Init(token string) error {
	d.authToken = base64.StdEncoding.EncodeToString([]byte(token))
	return nil
}

func (d *DynDNS) Resolve(fqdn string) (string, error) {
	return net.Resolve(fqdn)
}

func (d *DynDNS) SetApiEndpoint(ep string) {
	d.endpoint = ep
}

func (d *DynDNS) Update(fqdn, ip string) error {
	if d.authToken == "" {
		return fmt.Errorf("no authorization credentials found")
	}

	var (
		req *http.Request
		res *http.Response
		err error
	)

	log := slog.Default().With("endpoint", d.endpoint, "fqdn", fqdn)

	if req, err = http.NewRequest("GET", d.endpoint+"/nic/update", nil); err != nil {
		return fmt.Errorf("initialize HTTP connection to %s failed: %w", d.endpoint, err)
	}
	req.Header.Add("Authorization", "Basic "+d.authToken)
	req.Header.Add("User-Agent", strings.Join(
		[]string{"ddflare", version.Version, "dev@foggy.day"}, " "))

	q := req.URL.Query()
	q.Add("hostname", fqdn)
	if ip != "" {
		q.Add("myip", ip)
	}
	req.URL.RawQuery = q.Encode()
	if res, err = http.DefaultClient.Do(req); err != nil {
		return fmt.Errorf("connection to %s failed: %w", d.endpoint, err)
	}
	defer res.Body.Close()

	log.Debug("endpoint connected", "status", res.Status, "code", res.StatusCode)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("endpoint %q returned %d (%s) status", d.endpoint, res.StatusCode, res.Status)
	}

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil || len(body) == 0 {
		return fmt.Errorf("failure reading endpoint %q reply: %w", d.endpoint, err)
	}
	log.Debug("parsing reply message", "body", string(body))

	updateOk, msg := interpretResponse(string(body))

	if updateOk {
		log.Debug("update successful", "address", msg)
		return nil
	}

	return errors.New(msg)
}

// interpretResponse decodes the returned status messages and reports back to the caller:
//   - if the update was successful
//   - the arguments received (update successful) or an error message (in case of failed update)
func interpretResponse(resp string) (bool, string) {
	// let's ensure we don't get an empty string... we check in the caller, but better
	// stay safe for any future change may happen in the code
	if resp == "" {
		return false, "invalid return status"
	}

	respSlice := strings.Fields(resp)
	respLen := len(respSlice)
	if respLen > 2 {
		slog.Warn("unexpected number of arguments in reply", "reply", respSlice)
	}
	status := respSlice[0]
	msg := ""
	if respLen == 2 {
		msg = respSlice[1]
	}

	success := false
	switch status {
	case "good":
		fallthrough
	case "nochg":
		success = true
		if msg == "" {
			slog.Warn("invalid reply, missing argument", "status", status)
		}
	case "nohost":
		msg = "hostname supplied does not exist under specified account"
	case "badauth":
		msg = "invalid username password combination"
	case "badagent":
		msg = "client disabled"
	case "!donator":
		msg = "feature not available"
	case "abuse":
		msg = "username is blocked due to abuse"
	case "911":
		msg = "server Side fatal error: retry no sooner than 30 minutes"
	default:
		msg = "unknown status received"
	}
	return success, msg
}
