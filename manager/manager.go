package manager

import (
	"context"
	"sync"

	"github.com/gosom/go-gdc/entities"
)

type Crawler interface {
	Start(ctx context.Context, in <-chan entities.Job, out chan<- entities.Output)
}

type Manager struct {
	in       <-chan entities.Job
	crawlers []Crawler
}

func NewManager(in <-chan entities.Job, crawlers []Crawler) *Manager {
	ans := Manager{
		in:       in,
		crawlers: crawlers,
	}
	return &ans
}

func (o *Manager) Run(ctx context.Context) (<-chan entities.Output, <-chan error) {
	outc := make(chan entities.Output)
	errc := make(chan error, 1)
	go func() {
		defer close(outc)
		defer close(errc)
		wg := &sync.WaitGroup{}
		for i := range o.crawlers {
			wg.Add(1)
			go func(w Crawler) {
				defer wg.Done()
				w.Start(ctx, o.in, outc)
			}(o.crawlers[i])
		}
		wg.Wait()
	}()
	return outc, errc
}
