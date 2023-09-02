package cmd

import (
	"backuputil/common"
	"backuputil/restore"
	"github.com/spf13/cobra"
	"log"
	"os/exec"
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
		case "finance":
			restore.Finance(client)
		case "keycloak":
			restore.Keycloak(client)
		default:
			log.Fatal("Error: Namespace not supported")
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := exec.Command("argocd", "app", "list").Run(); err != nil {
			log.Fatal("error: argocd cli not configured.")
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
