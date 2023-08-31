package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var backupCmd = &cobra.Command{
	Use:   "backup [namespace]",
	Short: "Backs up services in a namespace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch strings.ToLower(args[0]) {
		default:
			log.Fatal("Error: Namespace not supported")
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
