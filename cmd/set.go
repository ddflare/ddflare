/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/fgiudici/ddflare/pkg/ddns"
	"github.com/fgiudici/ddflare/pkg/net"
	"github.com/spf13/cobra"
)

func newSetCommand() *cobra.Command {
	var ip, token *string

	cmd := &cobra.Command{
		Use:   "set {dnsName}",
		Short: "updates the dnsName domain A record to the ipAddress passed as argument",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		Long: `updates dnsName to the ip address passed with the --ip flag (or will use the current)
public ip if --ip is missing). The --token flag is required to authenticate to the
Cloudflare backend.
Example usage:
	ddflare set -t 7rvDyd2i3AqwesPtR_3wWWIoNNiGeoBmKQoiuyKj host.example.com`,

		RunE: func(cmd *cobra.Command, args []string) error {
			dnsName := args[0]
			ipAdd := *ip

			if ipAdd == "" {
				var err error
				if ipAdd, err = net.GetMyPub(); err != nil {
					return err
				}
			}

			domain := strings.Split(dnsName, ".")
			if len(domain) < 2 {
				return fmt.Errorf("%q is not a valid dns name", dnsName)
			}
			zoneName := domain[len(domain)-2] + "." + domain[len(domain)-1]

			ddns := ddns.Cloudflare{}
			if err := ddns.New(*token); err != nil {
				return err
			}

			if err := ddns.Write(dnsName, zoneName, ipAdd); err != nil {
				return err
			}
			return nil
		},
	}

	ip = cmd.Flags().StringP("ip", "i", "", "ip address")
	token = cmd.Flags().StringP("token", "t", "", "token")
	if err := cmd.MarkFlagRequired("token"); err != nil {
		log.Fatal(err)
	}

	return cmd
}
