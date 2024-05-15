/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gomodules.xyz/flags"
	v "gomodules.xyz/x/version"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "solrdump",
		Short: "Backup restore solr",
		Long:  `Command line tool to perform backup restore for solr`,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			flags.LoggerOptions = flags.GetOptions(c.Flags())
		},
	}
	rootCmd.AddCommand(v.NewCmdVersion())
	rootCmd.AddCommand(NewRunCmd())
	return rootCmd
}
