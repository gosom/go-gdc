package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/cobra"

	"github.com/gosom/go-gdc/crawler"
	"github.com/gosom/go-gdc/entities"
	"github.com/gosom/go-gdc/finisher"
	"github.com/gosom/go-gdc/manager"
	"github.com/gosom/go-gdc/parser"
	"github.com/gosom/go-gdc/repository"
)

func init() {
	var (
		regNum  string
		workers int
		format  string
		dsn     string
		refresh bool
	)
	defaultDsn := "host=localhost port=5432 dbname=gdc user=postgres password=password pool_max_conns=100"
	scraperCmd := &cobra.Command{
		Use:   "registration-number",
		Short: "Command line scraper",
		Long:  "Command line scraper",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var repo *repository.IndividualRepo
			var db *pgxpool.Pool
			defer func() {
				if db != nil {
					db.Close()
				}
			}()

			if refresh {
				var err error
				db, err = pgxpool.Connect(ctx, dsn)
				if err != nil {
					panic(err)
				}
				if err := db.Ping(ctx); err != nil {
					panic(err)
				}

				repo, err = repository.NewIndividualRepo(db)
				if err != nil {
					panic(err)
				}
			}

			var printer entities.ResultProcessor
			switch format {
			case "standard":
				printer = &finisher.SimpleStdout{}
			case "json":
				printer = &finisher.SimpleJson{}
			case "db":
				if !refresh {
					panic("Use also --refresh flag")
				}
				printer = finisher.NewDbSaver(repo)
			default:
				panic("unsupported format")
			}
			var jobs chan entities.Job
			p := parser.NewSingleResultParser()
			if len(regNum) > 0 {
				jobs = make(chan entities.Job, 1)
				singleJob := entities.NewSearchRegistrationNumberJob(p, regNum)
				jobs <- singleJob
			} else {
				items, err := repo.Select(ctx, nil)
				if err != nil {
					panic(err)
				}
				jobs = make(chan entities.Job, len(items))
				for i := range items {
					singleJob := entities.NewSearchRegistrationNumberJob(p, items[i].RegistrationNumber)
					jobs <- singleJob

				}
			}
			close(jobs)
			var netClient = &http.Client{
				Timeout: time.Second * 10,
			}
			var workhorses []manager.Crawler
			for i := 0; i < workers; i++ {
				cr := crawler.NewCrawler(netClient)
				workhorses = append(workhorses, cr)
			}

			mngr := manager.NewManager(jobs, workhorses)
			scraped, errc := mngr.Run(context.Background())

			printer.Run(scraped)

			if err := <-errc; err != nil {
				panic(err)
			}
		},
	}

	scraperCmd.Flags().StringVar(&regNum, "regNum", "", "Get details by registration number")
	scraperCmd.Flags().IntVar(&workers, "workers", 8, "Number of workers")
	scraperCmd.Flags().StringVar(&format, "format", "standard", "Output Format")
	scraperCmd.Flags().StringVar(&dsn, "dsn", defaultDsn, "database connection string")
	scraperCmd.Flags().BoolVar(&refresh, "refresh", false, "to refresh existing db")

	RootCmd.AddCommand(scraperCmd)
}
