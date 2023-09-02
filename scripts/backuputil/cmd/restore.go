package cmd

import (
	"backuputil/common"
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
		client, err := common.InitClient()
		if err != nil {
			log.Fatal(err)
		}

		switch strings.ToLower(args[0]) {
		case "immich":
			restore.Immich(client)
			break
		case "finance":
			restore.Finance(client)
			break
		default:
			log.Fatal("Error: Namespace not supported")
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
