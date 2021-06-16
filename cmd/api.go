package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gosom/go-gdc/server"
)

func init() {
	var (
		bind    string
		workers int
	)
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Simple REST API",
		Long:  "Simple REST API",
		Run: func(cmd *cobra.Command, args []string) {
			srv := server.NewServer(bind, workers)
			if err := srv.Start(context.Background()); err != nil {
				panic(err)
			}
		},
	}

	apiCmd.Flags().StringVar(&bind, "bind", ":8000", "Server bind address")
	apiCmd.Flags().IntVar(&workers, "workers", 8, "Number of workers")

	RootCmd.AddCommand(apiCmd)
}
