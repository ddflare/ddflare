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
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/fgiudici/ddflare/pkg/cflare"
	"github.com/fgiudici/ddflare/pkg/ddman"
	"github.com/fgiudici/ddflare/pkg/dyndns"
	"github.com/fgiudici/ddflare/pkg/net"
	"github.com/urfave/cli/v2"
)

const (
	IPADDR   = "DDFLARE_IP_ADDRESS"
	TOKEN    = "DDFLARE_API_TOKEN"
	INTERVAL = "DDFLARE_CHECK_INTERVAL"
	SVC      = "DDFLARE_SERVICE_PROVIDER"
	USER     = "DDFLARE_USER"
	PASSWD   = "DDFLARE_PASSWORD"
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
				EnvVars: []string{IPADDR},
			},
			&cli.StringFlag{
				Name:    "api-token",
				Aliases: []string{"t"},
				Usage:   "API authentication token",
				EnvVars: []string{TOKEN},
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
				Usage:   "interval between consecutive checks (implies --check)",
				EnvVars: []string{INTERVAL},
			},
			&cli.BoolFlag{
				Name:    "loop",
				Aliases: []string{"l"},
				Usage:   "shorthand for --check --interval 5m",
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "svc",
				Aliases: []string{"s"},
				Usage:   "DDNS service provider [cflare, noip, $URL]",
				EnvVars: []string{SVC},
				Value:   "cflare",
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "username (alternative to the 'api-token')",
				EnvVars: []string{USER},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "password (alternative to the 'api-token')",
				EnvVars: []string{PASSWD},
			},
		},
		Action: func(cCtx *cli.Context) error {
			var (
				conf *setConf
				err  error
			)
			if conf, err = newSetConf(cCtx); err != nil {
				cli.ShowSubcommandHelp(cCtx)
				return err
			}

			var ddns ddman.DNSManager
			if conf.svc == "cflare" {
				ddns = cflare.New()
			} else {
				ddns = dyndns.New(conf.svc)
			}

			if err = ddns.Init(conf.token); err != nil {
				slog.Error("DNS Manager initialization failed", "error", err)
				return err
			}

			for {
				if err = updateFQDN(ddns, conf); err != nil {
					slog.Error("FQDN update failed", "fqdn", conf.fqdn, "ip", conf.address, "error", err)
					return err
				}
				slog.Info("FQDN update successful", "fqdn", conf.fqdn, "ip", conf.address)

				if conf.interval == 0 {
					return nil
				}
				time.Sleep(conf.interval)
			}
		},
	}

	return cmd
}

type setConf struct {
	fqdn           string
	address        string
	token          string
	check          bool
	interval       time.Duration
	loop           bool
	svc            string
	lastSetAddress string
}

func newSetConf(cCtx *cli.Context) (*setConf, error) {
	conf := &setConf{}

	conf.fqdn = cCtx.Args().First()
	if conf.fqdn == "" {
		return nil, errors.New("'fqdn' arg is missing")
	}

	conf.token = cCtx.String("api-token")
	if conf.token == "" {
		user := cCtx.String("user")
		passwd := cCtx.String("password")
		if user == "" || passwd == "" {
			return nil, errors.New("auth credential missing ('api-token' or 'user' + 'password')")
		}
		conf.token = user + ":" + passwd
	}

	conf.address = cCtx.String("address")
	if conf.address == "" {
		var err error
		if conf.address, err = net.GetMyPub(); err != nil {
			return nil, err
		}
	}
	conf.check = cCtx.Bool("check")
	conf.interval = cCtx.Duration("interval")
	conf.loop = cCtx.Bool("loop")
	if conf.loop && conf.interval == time.Duration(0) {
		conf.interval = 5 * time.Minute
	}
	if conf.interval > time.Duration(0) {
		conf.check = true
	}

	conf.svc = cCtx.String("svc")
	switch conf.svc {
	case "noip":
		conf.svc = "https://dynupdate.no-ip.com"
	case "ddns":
		conf.svc = "https://update.ddns.org"
	}
	return conf, nil
}

func updateFQDN(ddns ddman.DNSManager, conf *setConf) error {
	address := conf.address
	if address == "" {
		var err error
		if address, err = net.GetMyPub(); err != nil {
			return fmt.Errorf("cannot retrieve public address: %w", err)
		}
		slog.Debug("Got public IP", "address", address)
	}
	if conf.check {
		if address == conf.lastSetAddress {
			return nil
		}
		conf.lastSetAddress = address
		if isFQDNUpToDate(ddns, conf.fqdn, address) {
			slog.Debug("FQDN is up to date", "fqdn", conf.fqdn, "ip", conf.address)
			return nil
		}
	}

	if err := ddns.Update(conf.fqdn, address); err != nil {
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
