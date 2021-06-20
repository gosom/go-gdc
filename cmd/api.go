package cmd

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"

	"github.com/gosom/go-gdc/repository"
	"github.com/gosom/go-gdc/server"
	"github.com/gosom/go-gdc/utils"
)

func init() {
	bind := utils.GetEnv("BIND_ADDRESS", ":8000")
	workers, err := strconv.ParseInt(utils.GetEnv("WORKERS", "16"), 10, 64)
	if err != nil {
		panic(err)
	}
	defaultDsn := "host=localhost port=5432 dbname=gdc user=postgres password=password pool_max_conns=100"
	dsn := utils.GetEnv("DSN", defaultDsn)
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Simple REST API",
		Long:  "Simple REST API",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			db, err := pgxpool.Connect(ctx, dsn)
			if err != nil {
				panic(err)
			}
			if err := db.Ping(ctx); err != nil {
				panic(err)
			}
			defer db.Close()

			repo, err := repository.NewIndividualRepo(db)
			if err != nil {
				panic(err)
			}
			srv := server.NewServer(bind, int(workers), repo)
			if err := srv.Start(context.Background()); err != nil {
				panic(err)
			}
		},
	}

	RootCmd.AddCommand(apiCmd)
}
