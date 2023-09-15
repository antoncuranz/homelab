package cmd

import (
	"backuputil/common"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"
	"log"
	"strings"
	"time"
)

var backupCmd = &cobra.Command{
	Use:   "backup [namespace]",
	Short: "Backs up services in a namespace",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := common.InitClient()
		if err != nil {
			log.Fatal(err)
		}

		timestamp := time.Now().Format("200601021504")
		k8upNamespaces := []string{"immich", "homebridge", "pihole", "servarr", "notifications", "paperless"}
		dbNamespaces := []string{"immich", "finance", "keycloak", "paperless"}

		if len(args) == 0 {
			fmt.Println("Backing up database...")
			if err := common.CreateDatabaseBackup(client, "manual-"+timestamp); err != nil {
				log.Fatal(err)
			}

			fmt.Println("Backing up pvs...")
			for _, namespace := range k8upNamespaces {
				if err := common.CreateBackup(client, namespace, "manual-"+namespace+"-"+timestamp); err != nil {
					log.Fatal(err)
				}
			}

			return
		}

		namespace := strings.ToLower(args[0])
		if slices.Contains([]string{"db", "database"}, namespace) {
			fmt.Println("Backing up database...")
			if err := common.CreateDatabaseBackup(client, "manual-"+timestamp); err != nil {
				log.Fatal(err)
			}
		} else if slices.Contains(k8upNamespaces, namespace) {
			if slices.Contains(dbNamespaces, namespace) {
				fmt.Println("Backing up database...")
				if err := common.CreateDatabaseBackup(client, "manual-"+timestamp); err != nil {
					log.Fatal(err)
				}
			}

			fmt.Println("Backing up pv...")
			if err := common.CreateBackup(client, namespace, "manual-"+namespace+"-"+timestamp); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("Error: Namespace not supported")
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
