/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	cmds "github.com/pritamdas99/solr-dump/pkg/cmd"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
	"k8s.io/klog/v2"
)

func main() {
	rootCmd := cmds.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		klog.Warning(err)
	}

}
