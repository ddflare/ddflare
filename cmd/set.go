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
	"log/slog"
	"time"

	"github.com/fgiudici/ddflare/pkg/cflare"
	"github.com/fgiudici/ddflare/pkg/ddman"
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
				Name:    "address",
				Aliases: []string{"a"},
				Usage:   "IP address to set (current public address if not specified)",
				EnvVars: []string{"IPADDR"},
			},
			&cli.StringFlag{
				Name:     "api-token",
				Aliases:  []string{"t"},
				Usage:    "API authentication token",
				EnvVars:  []string{"TOKEN"},
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "check",
				Aliases: []string{"c"},
				Usage:   "check if the record needs actual update before writing",
				Value:   false,
			},
			&cli.DurationFlag{
				Name:    "interval",
				Aliases: []string{"i"},
				Usage:   "interval to wait between consecutive checks (implies --check)",
				EnvVars: []string{"INTERVAL"},
			},
			&cli.BoolFlag{
				Name:    "loop",
				Aliases: []string{"l"},
				Usage:   "shorthand for --check --interval 5m",
				Value:   false,
			},
		},
		Action: func(cCtx *cli.Context) error {
			fqdn := cCtx.Args().First()
			ipAdd := cCtx.String("address")
			check := cCtx.Bool("check")
			interval := cCtx.Duration("interval")
			loop := cCtx.Bool("loop")
			token := cCtx.String("api-token")
			var err error

			if ipAdd == "" {
				if ipAdd, err = net.GetMyPub(); err != nil {
					return err
				}
			}

			if loop {
				if interval == time.Duration(0) {
					interval = 5 * time.Minute
				}
			}

			if interval > time.Duration(0) {
				check = true
			}

			// cflare is the only backend right now
			var ddns ddman.DNSManager = cflare.New()
			if err = ddns.Init(token); err != nil {
				slog.Error("DNS Manager initialization failed", "error", err)
				return err
			}

			for {
				if err = updateFQDN(ddns, fqdn, ipAdd, check); err != nil {
					slog.Error("FQDN update failed", "fqdn", fqdn, "ip", ipAdd, "error", err)
				} else {
					slog.Info("FQDN update successful", "fqdn", fqdn, "ip", ipAdd)
				}

				if interval == 0 {
					return err
				}
				time.Sleep(interval)
			}
		},
	}

	return cmd
}

func updateFQDN(ddns ddman.DNSManager, fqdn, ipAdd string, check bool) error {
	if check && isFQDNUpToDate(ddns, fqdn, ipAdd) {
		slog.Debug("FQDN is up to date", "fqdn", fqdn, "ip", ipAdd)
		return nil
	}

	if err := ddns.Update(fqdn, ipAdd); err != nil {
		return err
	}

	return nil
}

func isFQDNUpToDate(ddns ddman.DNSManager, fqdn, ipAdd string) bool {
	var (
		resIp string
		err   error
	)
	if resIp, err = ddns.Resolve(fqdn); err != nil {
		slog.Error(err.Error())
		return false
	}
	if ipAdd != resIp {
		slog.Debug("FQDN requires update", "fqdn", fqdn, "ip", resIp)
		return false
	}
	return true
}
