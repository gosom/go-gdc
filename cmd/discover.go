package cmd

import (
	"context"
	"net/http"
	"time"
	//"fmt"

	"github.com/spf13/cobra"

	"github.com/gosom/go-gdc/crawler"
	"github.com/gosom/go-gdc/entities"
	"github.com/gosom/go-gdc/finisher"
	"github.com/gosom/go-gdc/manager"
	"github.com/gosom/go-gdc/parser"
	"github.com/gosom/go-gdc/utils"
)

func init() {
	var (
		postcodesFname string
		workers        int
	)

	discoverCmd := &cobra.Command{
		Use:   "discover",
		Short: "Scapes all registration numbers in the postcodes provided",
		Long:  "Scapes all registration numbers in the postcodes provided",
		Run: func(cmd *cobra.Command, args []string) {
			if postcodesFname == "" {
				panic("provide postcodes")
			}
			postcodes, err := utils.ReadLines(postcodesFname)
			if err != nil {
				panic(err)
			}
			p := parser.NewListinResultParser()
			jobs := make(chan entities.Job, len(postcodes))
			for i := range postcodes {
				job := entities.NewDiscoverJob(p, postcodes[i], "POST", "")
				jobs <- job
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

			var printer entities.ResultProcessor
			printer = &finisher.SimpleStdout{}

			mngr := manager.NewManager(jobs, workhorses)
			scraped, errc := mngr.Run(context.Background())

			printer.Run(scraped)

			if err := <-errc; err != nil {
				panic(err)
			}

		},
	}

	discoverCmd.Flags().StringVar(&postcodesFname, "postcodes", "", "Provide path to file which contains one postcode per line")
	discoverCmd.Flags().IntVar(&workers, "workers", 8, "Number of workers")

	RootCmd.AddCommand(discoverCmd)
}
