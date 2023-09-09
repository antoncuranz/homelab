package cmd

import (
	"backuputil/common"
	"backuputil/restore"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [namespace]",
	Short: "Restores services in a namespace",
	Args:  cobra.ExactArgs(1), // TODO: maximum 1 - if 0, restore everything (db last)!
	Run: func(cmd *cobra.Command, args []string) {
		client, err := common.InitClient()
		if err != nil {
			log.Fatal(err)
		}

		snapshotMap, err := common.CreateResticSnapshotMap()
		if err != nil {
			fmt.Println("Error fetching restic snapshots")
			log.Fatal(err)
		}

		namespace := strings.ToLower(args[0])
		namespacedSnapshotMap := snapshotMap[namespace]
		switch namespace {
		case "db":
			fallthrough
		case "database":
			restore.Database(client)
		case "immich":
			restore.Immich(client, namespacedSnapshotMap)
		case "homebridge":
			restore.Homebridge(client, namespacedSnapshotMap)
		case "notifications":
			restore.Notifications(client, namespacedSnapshotMap)
		case "paperless":
			restore.Paperless(client, namespacedSnapshotMap)
		case "servarr":
			restore.Servarr(client, namespacedSnapshotMap)
		default:
			log.Fatal("Error: Namespace not supported")
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(os.Getenv("RESTIC_PASSWORD")) == 0 {
			log.Fatal("error: RESTIC_PASSWORD environment variable must be set.")
		}
		if len(os.Getenv("AWS_SECRET_ACCESS_KEY")) == 0 {
			log.Fatal("error: AWS_SECRET_ACCESS_KEY environment variable must be set.")
		}
		if len(os.Getenv("AWS_ACCESS_KEY_ID")) == 0 {
			log.Fatal("error: AWS_ACCESS_KEY_ID environment variable must be set.")
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
