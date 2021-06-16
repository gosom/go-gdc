package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gosom/go-gdc/crawler"
	"github.com/gosom/go-gdc/entities"
	"github.com/gosom/go-gdc/manager"
	"github.com/gosom/go-gdc/parser"
)

type Server struct {
	s        *http.Server
	mngr     *manager.Manager
	jobs     chan entities.Job
	registry map[string]chan entities.Output
	lock     *sync.Mutex
}

func NewServer(bind string, workers int) *Server {
	ans := Server{
		jobs:     make(chan entities.Job),
		lock:     &sync.Mutex{},
		registry: make(map[string]chan entities.Output),
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/registration-number-search",
		searchByRegistrationNumber(ans.jobs, ans.registry, ans.lock),
	)

	ans.s = &http.Server{
		Addr:           bind,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	var workhorses []manager.Crawler
	for i := 0; i < workers; i++ {
		cr := crawler.NewCrawler(netClient)
		workhorses = append(workhorses, cr)
	}
	ans.mngr = manager.NewManager(ans.jobs, workhorses)
	return &ans
}

func (o *Server) Start(ctx context.Context) error {
	scraped, errc := o.mngr.Run(ctx)
	go func() {
		for el := range scraped {
			o.lock.Lock()
			if ch, ok := o.registry[el.Job.GetID()]; ok {
				ch <- el
			}
			o.lock.Unlock()
		}
	}()
	go func() {
		if err := <-errc; err != nil {
			panic(err)
		}
	}()
	return o.s.ListenAndServe()
}

func searchByRegistrationNumber(q chan entities.Job, registry map[string]chan entities.Output,
	lock *sync.Mutex) http.HandlerFunc {
	p := parser.NewSingleResultParser()
	return func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["regNum"]
		if !ok || len(keys[0]) < 1 {
			msg := map[string]string{
				"error": "Url Param 'regNum' is missing'",
			}
			renderJson(w, http.StatusBadRequest, msg)
			return
		}
		job := entities.NewSearchRegistrationNumberJob(p, keys[0])
		ch := make(chan entities.Output, 1)
		defer func() {
			close(ch)
			lock.Lock()
			delete(registry, job.GetID())
			lock.Unlock()
		}()

		lock.Lock()
		registry[job.GetID()] = ch
		lock.Unlock()
		q <- job
		result := <-ch
		renderJson(w, http.StatusOK, result)
	}
}

func renderJson(w http.ResponseWriter, status int, body interface{}) {
	var js []byte
	if body != nil {
		var err error
		js, err = json.Marshal(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
