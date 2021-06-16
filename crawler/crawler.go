package crawler

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gosom/go-gdc/entities"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.157 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/54.0",
}

type Crawler struct {
	client *http.Client
}

func NewCrawler(client *http.Client) *Crawler {
	ans := Crawler{
		client: client,
	}
	return &ans
}

func (o *Crawler) Start(ctx context.Context, in <-chan entities.Job, out chan<- entities.Output) {
	for {
		select {
		case <-ctx.Done():
			return
		case j, ok := <-in:
			if !ok {
				return
			}
			out <- o.scrape(ctx, j)
		}
	}
}

func (o *Crawler) scrape(ctx context.Context, job entities.Job) entities.Output {
	ans := entities.Output{
		Job: job,
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := o.prepareReq(reqCtx, job)
	if err != nil {
		ans.Error = err
		return ans
	}
	resp, err := o.client.Do(req)
	if err != nil {
		ans.Error = err
		return ans
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
	ans.StatusCode = resp.StatusCode
	if ans.StatusCode == 200 {
		ans.Body, ans.Error = io.ReadAll(resp.Body)
		if ans.Error != nil {
			return ans
		}
		parser := job.GetParser()
		ans.Individual, ans.Error = parser.Parse(ctx, ans.Body)
	}
	return ans
}

func (o *Crawler) prepareReq(ctx context.Context, job entities.Job) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, job.GetUrl(), nil)
	if err != nil {
		return req, err
	}
	n := rand.Int() % len(userAgents)
	req.Header.Add("User-Agent", userAgents[n])
	req.Header.Add("Cookie", "GDC_cookieconsent_status=deny")
	return req, nil
}
