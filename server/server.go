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
	mux.HandleFunc("/postcode-search",
		searchByPostcode(ans.jobs, ans.registry, ans.lock),
	)

	ans.s = &http.Server{
		Addr:           bind,
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
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

type PostCodeResponse struct {
	Postcode    string
	Count       int
	Individuals []entities.Individual
	Errors      []string
}

func searchByPostcode(q chan entities.Job, registry map[string]chan entities.Output,
	lock *sync.Mutex) http.HandlerFunc {
	p := parser.NewListinResultParser()
	single := parser.NewSingleResultParser()
	return func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["postcode"]
		if !ok || len(keys[0]) < 1 {
			msg := map[string]string{
				"error": "Url Param 'postcode' is missing'",
			}
			renderJson(w, http.StatusBadRequest, msg)
			return
		}
		job := entities.NewDiscoverJob(p, keys[0], "POST", "")
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
		items := <-ch

		var regNumbers []string
		for i := range items.Individuals {
			regNumbers = append(regNumbers, items.Individuals[i].RegistrationNumber)
		}
		result := PostCodeResponse{
			Postcode: keys[0],
		}

		if len(regNumbers) > 0 {
			other := make(chan entities.Output, len(regNumbers))
			wg := &sync.WaitGroup{}
			for i := range regNumbers {
				wg.Add(1)
				go func(num string) {
					innerJob := entities.NewSearchRegistrationNumberJob(single, num)
					innerCh := make(chan entities.Output, 1)
					defer func() {
						close(innerCh)
						lock.Lock()
						delete(registry, innerJob.GetID())
						lock.Unlock()
						wg.Done()
					}()
					lock.Lock()
					registry[innerJob.GetID()] = innerCh
					lock.Unlock()
					q <- innerJob
					innerResult := <-innerCh
					other <- innerResult
				}(regNumbers[i])
			}
			wg.Wait()
			close(other)
			for el := range other {
				if el.Error != nil {
					result.Errors = append(result.Errors, el.Error.Error())
				} else {
					result.Individuals = append(result.Individuals, el.Individuals...)
				}
			}
			result.Count = len(result.Individuals)
		}

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
