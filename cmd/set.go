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

package cmd

import (
	"fmt"
	"strings"

	"github.com/fgiudici/ddflare/pkg/cflare"
	"github.com/fgiudici/ddflare/pkg/ddns"
	"github.com/fgiudici/ddflare/pkg/net"
	"github.com/urfave/cli/v2"
)

func newSetCommand() *cli.Command {
	cmd := &cli.Command{
		Name:      "set",
		Usage:     "updates the A record of the fqdn passed as argument",
		Args:      true,
		ArgsUsage: "fqdn",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ip",
				Aliases: []string{"i"},
				Usage:   "ip address (current public IP if not specified)",
				EnvVars: []string{"IPADDRESS"},
			},
			&cli.StringFlag{
				Name:     "token",
				Aliases:  []string{"t"},
				Usage:    "token",
				EnvVars:  []string{"TOKEN"},
				Required: true,
			},
		},
		Action: func(cCtx *cli.Context) error {
			fqdn := cCtx.Args().First()
			ipAdd := cCtx.String("ip")
			token := cCtx.String("token")
			var (
				err      error
				zoneName string
			)

			if ipAdd == "" {
				if ipAdd, err = net.GetMyPub(); err != nil {
					return err
				}
			}

			if zoneName, err = getZone(fqdn); err != nil {
				return err
			}

			// cflare is the only backend right now
			var ddns ddns.Recorder = cflare.New()

			if err := ddns.Init(token); err != nil {
				return err
			}

			if err := ddns.Write(fqdn, zoneName, ipAdd); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}

func getZone(fqdn string) (string, error) {
	domain := strings.Split(fqdn, ".")
	if len(domain) < 2 {
		return "", fmt.Errorf("%q is not a valid dns name", fqdn)
	}
	zone := domain[len(domain)-2] + "." + domain[len(domain)-1]
	return zone, nil
}
