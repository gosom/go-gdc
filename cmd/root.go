package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "go-gdc",
	Short: "General Dental Council Scraper",
	Long:  "General Dental Council Scraper",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
