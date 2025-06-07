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
	"log/slog"
	"time"

	"github.com/ddflare/ddflare"
	"github.com/ddflare/ddflare/pkg/version"
	"github.com/urfave/cli/v2"
)

const (
	USERAGENT = "ddflare-go-"
	IPADDR    = "DDFLARE_IP_ADDRESS"
	TOKEN     = "DDFLARE_API_TOKEN"
	INTERVAL  = "DDFLARE_CHECK_INTERVAL"
	SVC       = "DDFLARE_SERVICE_PROVIDER"
	USER      = "DDFLARE_USER"
	PASSWD    = "DDFLARE_PASSWORD"
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
			&cli.DurationFlag{
				Name:    "interval",
				Aliases: []string{"i"},
				Usage:   "interval between consecutive updates",
				EnvVars: []string{INTERVAL},
			},
			&cli.BoolFlag{
				Name:    "loop",
				Aliases: []string{"l"},
				Usage:   "shorthand for --interval 5m",
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "svc",
				Aliases: []string{"s"},
				Usage:   "DDNS service provider [cflare, dyn, noip, ddns, $URL]",
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
			dm := conf.dm

			for {
				ip := conf.address
				if ip == "" {
					if ip, err = ddflare.GetPublicIP(); err != nil {
						slog.Error("IP Public retrieval failed", "error", err)
						return err
					}
				}
				if err = dm.UpdateFQDN(conf.fqdn, ip); err != nil {
					slog.Error("FQDN update failed", "fqdn", conf.fqdn, "ip", ip, "error", err)
					return err
				}
				slog.Info("FQDN update successful", "fqdn", conf.fqdn, "ip", ip)

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
	fqdn     string
	address  string
	interval time.Duration
	loop     bool
	dm       *ddflare.DNSManager
}

func newSetConf(cCtx *cli.Context) (*setConf, error) {
	conf := &setConf{}

	conf.fqdn = cCtx.Args().First()
	if conf.fqdn == "" {
		return nil, errors.New("'fqdn' arg is missing")
	}

	svc := cCtx.String("svc")
	switch svc {
	case "cflare":
		conf.dm, _ = ddflare.NewDNSManager(ddflare.Cloudflare)
	case "dyn":
		conf.dm, _ = ddflare.NewDNSManager(ddflare.Dyn)
	case "noip":
		conf.dm, _ = ddflare.NewDNSManager(ddflare.NoIP)
	case "ddns":
		conf.dm, _ = ddflare.NewDNSManager(ddflare.DDNS)
	default:
		conf.dm, _ = ddflare.NewDNSManager(ddflare.DDNS)
		conf.dm.SetApiEndpoint(svc)
	}
	token := cCtx.String("api-token")
	if token == "" {
		user := cCtx.String("user")
		passwd := cCtx.String("password")
		if user == "" || passwd == "" {
			return nil, errors.New("auth credential missing ('api-token' or 'user' + 'password')")
		}
		token = user + ":" + passwd
	}
	if err := conf.dm.Init(token); err != nil {
		return nil, errors.New("DNS Manager auth initialization failed")
	}

	conf.dm.SetUserAgent(USERAGENT + version.Version)

	conf.address = cCtx.String("address")
	conf.interval = cCtx.Duration("interval")
	conf.loop = cCtx.Bool("loop")
	if conf.loop && conf.interval == time.Duration(0) {
		conf.interval = 5 * time.Minute
	}

	return conf, nil
}
