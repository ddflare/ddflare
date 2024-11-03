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

	"github.com/fgiudici/ddflare/pkg/net"
	"github.com/urfave/cli/v2"
)

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
			domain := cCtx.Args().First()
			if domain == "" {
				domain = "pubIp"
			}
			var ipAdd string
			var err error
			var quiet = cCtx.Bool("quiet")

			switch domain {
			case "pubIp":
				ipAdd, err = net.GetMyPub()
			default:
				ipAdd, err = net.Resolve(domain)
			}

			if err != nil {
				return err
			}
			if quiet {
				fmt.Printf("%s", ipAdd)
			} else {
				fmt.Printf("%s --> %s\n", domain, ipAdd)
			}
			return nil
		},
	}
	return cmd
}
