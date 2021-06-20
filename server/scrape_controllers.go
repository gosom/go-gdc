package server

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/gosom/go-gdc/entities"
	"github.com/gosom/go-gdc/parser"
)

func searchByRegistrationNumber(q chan entities.Job, registry map[string]chan entities.Output,
	lock *sync.Mutex) http.HandlerFunc {
	p := parser.NewSingleResultParser()
	return func(w http.ResponseWriter, r *http.Request) {
		regNum := chi.URLParam(r, "regNum")
		job := entities.NewSearchRegistrationNumberJob(p, regNum)
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
		postcode := chi.URLParam(r, "postcode")
		job := entities.NewDiscoverJob(p, postcode, "POST", "")
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
			Postcode: postcode,
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
