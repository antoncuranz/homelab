package cmd

import (
	"backuputil/common"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var backupCmd = &cobra.Command{
	Use:   "backup [namespace]",
	Short: "Backs up services in a namespace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := common.InitClient()
		if err != nil {
			log.Fatal(err)
		}

		namespace := strings.ToLower(args[0])
		switch namespace {
		case "immich":
			fallthrough
		case "keycloak":
			fallthrough
		case "finance":
			if err := common.CreateBackup(client, namespace, "backup-"+namespace); err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatal("Error: Namespace not supported")
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
