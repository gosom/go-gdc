package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/gosom/go-gdc/crawler"
	"github.com/gosom/go-gdc/entities"
	"github.com/gosom/go-gdc/finisher"
	"github.com/gosom/go-gdc/manager"
	"github.com/gosom/go-gdc/parser"
)

func init() {
	var (
		regNum  string
		fpath   string
		workers int
		format  string
	)
	scraperCmd := &cobra.Command{
		Use:   "registration-number",
		Short: "Command line scraper",
		Long:  "Command line scraper",
		Run: func(cmd *cobra.Command, args []string) {
			var printer entities.ResultProcessor
			switch format {
			case "standard":
				printer = &finisher.SimpleStdout{}
			case "json":
				printer = &finisher.SimpleJson{}
			}
			var jobs chan entities.Job
			p := parser.NewSingleResultParser()
			if len(regNum) > 0 {
				jobs = make(chan entities.Job, 1)
				singleJob := entities.NewSearchRegistrationNumberJob(p, regNum)
				jobs <- singleJob
			} else {
				// TODO
				panic("not implemented")
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
	scraperCmd.Flags().StringVar(&fpath, "fpath", "", "Provide path to file which contains one registration number per line")
	scraperCmd.Flags().IntVar(&workers, "workers", 8, "Number of workers")
	scraperCmd.Flags().StringVar(&format, "format", "standard", "Output Format")

	RootCmd.AddCommand(scraperCmd)
}
