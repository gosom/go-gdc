package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"

	"github.com/gosom/go-gdc/entities"
	"github.com/gosom/go-gdc/repository"
	"github.com/gosom/go-gdc/server"
)

func init() {
	var (
		fpath string
		dsn   string
	)
	insertCmd := &cobra.Command{
		Use:   "insert",
		Short: "Insert scraped data",
		Long:  "Insert scraped data",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_ = ctx
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

			items, err := readIndividuals(fpath)
			if err != nil {
				panic(err)
			}
			if err := repo.BulkInsert(ctx, items); err != nil {
				panic(err)
			}
		},
	}

	defaultDsn := "host=localhost port=5432 dbname=gdc user=postgres password=password pool_max_conns=100"
	insertCmd.Flags().StringVar(&dsn, "dsn", defaultDsn, "database connection string")
	insertCmd.Flags().StringVar(&fpath, "fpath", "", "file containing scraped data")

	RootCmd.AddCommand(insertCmd)
}

func readIndividuals(fname string) ([]entities.Individual, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var items []entities.Individual
	seen := make(map[string]bool)
	d := json.NewDecoder(f)
	for {
		var v server.PostCodeResponse
		if err := d.Decode(&v); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		for i := range v.Individuals {
			if !seen[v.Individuals[i].RegistrationNumber] {
				seen[v.Individuals[i].RegistrationNumber] = true
				items = append(items, v.Individuals[i])
			}
		}
	}
	return items, nil
}
