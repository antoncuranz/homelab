package cmd

import (
	"backuputil/restore"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [namespace]",
	Short: "Restores services in a namespace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch strings.ToLower(args[0]) {
		case "immich":
			restore.RestoreImmich()
			break
		default:
			log.Fatal("Error: Namespace not supported")
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
