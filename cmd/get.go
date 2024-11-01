/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/fgiudici/ddflare/pkg/net"
	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	var quiet *bool

	cmd := &cobra.Command{
		Use:   "get {domain}",
		Short: "retrieve the IP address of the target domain",
		Args:  cobra.MatchAll(cobra.ExactArgs(1)),
		Long: `resolves the host IP address of the domain passed as argument.
The argument "pubIp" is special and allows to retrieve the current public address`,
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]
			var ipAdd string
			var err error

			switch domain {
			case "pubIp":
				ipAdd, err = net.GetMyPub()
			default:
				ipAdd, err = net.Resolve(domain)
			}

			if err != nil {
				return err
			}
			if *quiet {
				fmt.Printf("%s", ipAdd)
			} else {
				fmt.Printf("%s --> %s\n", domain, ipAdd)
			}
			return nil
		},
	}

	quiet = cmd.Flags().BoolP("quiet", "q", false, "quiet mode")

	return cmd
}
