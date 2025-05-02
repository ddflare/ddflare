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
	"log/slog"

	"github.com/ddflare/ddflare/pkg/net"
	"github.com/urfave/cli/v2"
)

const pubIP = "PublicIP"

func newGetCommand() *cli.Command {
	cmd := &cli.Command{
		Name:  "get",
		Usage: "retrieve the IP address of the target domain",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Value:   false,
				Usage:   "quiet mode",
			},
		},
		Action: func(cCtx *cli.Context) error {
			fqdn := cCtx.Args().First()
			if fqdn == "" {
				fqdn = pubIP
			}
			var ipAdd string
			var err error
			var quiet = cCtx.Bool("quiet")

			switch fqdn {
			case pubIP:
				ipAdd, err = net.GetMyPub()
			default:
				ipAdd, err = net.Resolve(fqdn)
			}

			if err != nil {
				slog.Error("IP retrieval failed", "fqdn", fqdn, "error", err)
				return err
			}
			if quiet {
				fmt.Printf("%s", ipAdd)
			} else {
				fmt.Printf("%s: %s\n", fqdn, ipAdd)
			}
			return nil
		},
	}
	return cmd
}
