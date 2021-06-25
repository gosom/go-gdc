package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/gosom/go-gdc/crawler"
	"github.com/gosom/go-gdc/entities"
	"github.com/gosom/go-gdc/manager"
)

type IndividualRepo interface {
	GetByRegNum(ctx context.Context, regNum string) (entities.Individual, error)
	Select(ctx context.Context, conditions []entities.Condition) ([]entities.Individual, error)
	Search(ctx context.Context, term string) ([]entities.Individual, error)
}

type Server struct {
	s        *http.Server
	mngr     *manager.Manager
	jobs     chan entities.Job
	registry map[string]chan entities.Output
	lock     *sync.Mutex
	repo     IndividualRepo
}

func NewServer(bind string, workers int, repo IndividualRepo) *Server {
	ans := Server{
		jobs:     make(chan entities.Job),
		lock:     &sync.Mutex{},
		registry: make(map[string]chan entities.Output),
		repo:     repo,
	}
	mux := ans.getRouter()

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

func (o *Server) getRouter() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(120 * time.Second))

	mux.Get("/scrape/registration-number/{regNum}",
		searchByRegistrationNumber(o.jobs, o.registry, o.lock),
	)
	mux.Get("/scrape/postcode/{postcode}",
		searchByPostcode(o.jobs, o.registry, o.lock),
	)

	mux.Route("/individuals", func(r chi.Router) {
		r.Get("/", selectIndividuals(o.repo))
		r.Get("/{regNum}", getIndividualByRegNum(o.repo))
		r.Get("/search/{term}", searchIndividuals(o.repo))
	})

	return mux
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
