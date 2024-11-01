/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// This is called by main.main().
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "ddflare",
		Short: "update dns entries via cloudflare APIs",
	}

	rootCmd.AddCommand(
		newGetCommand(),
		newSetCommand(),
		newVersionCommand(),
	)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
