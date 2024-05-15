/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	solr_dump "github.com/pritamdas99/solr-dump/pkg/solr-dump"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var (
	action     string
	actions    = []string{"backup", "restore"}
	db         string
	namespace  string
	location   string
	repository string
	runCmd     = &cobra.Command{
		Use:   "run",
		Short: "Launch solr-dump",
		Run: func(cmd *cobra.Command, args []string) {
			dumper, err := solr_dump.NewSolrDump(action, db, namespace, location, repository)
			if err != nil {
				klog.Error(err)
			}
			dumper.Execute()
		},
	}
)

func NewRunCmd() *cobra.Command {
	return runCmd
}

func init() {
	runCmd.PersistentFlags().StringVarP(&action, "action", "a", "backup", fmt.Sprintf("The operation to carry out.\n\tSupported values are %v", actions))
	runCmd.PersistentFlags().StringVarP(&db, "db", "d", "", fmt.Sprintf("db instance to take backup"))
	runCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", fmt.Sprintf("Namespace of db instance"))
	runCmd.PersistentFlags().StringVarP(&location, "location", "l", "", fmt.Sprintf("location of cloud backend where backups will be stored"))
	runCmd.PersistentFlags().StringVarP(&repository, "repository", "r", "", fmt.Sprintf("repository of the backend"))
}
