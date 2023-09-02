package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "backuputil",
	Short: "Utility for backing up and restoring my homelab services",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if len(os.Getenv("RESTIC_PASSWORD")) == 0 {
			log.Fatal("error: RESTIC_PASSWORD environment variable must be set.")
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.backuputil.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
