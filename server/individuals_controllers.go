package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gosom/go-gdc/entities"
)

type SingleIndividual struct {
	entities.Individual
}

func getIndividualByRegNum(repo IndividualRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		regNum := chi.URLParam(r, "regNum")
		item, err := repo.GetByRegNum(r.Context(), regNum)
		if err != nil {
			renderJson(w, http.StatusNotFound, map[string]string{
				"error": err.Error(),
			},
			)
			return
		}
		ans := SingleIndividual{
			Individual: item,
		}
		renderJson(w, http.StatusOK, ans)
	}
}

func searchIndividuals(repo IndividualRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		term := chi.URLParam(r, "term")
		items, err := repo.Search(r.Context(), term)
		if err != nil {
			renderJson(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			},
			)
			return
		}
		ans := make([]SingleIndividual, 0, len(items))
		for i := range items {
			el := SingleIndividual{
				Individual: items[i],
			}
			ans = append(ans, el)
		}
		renderJson(w, http.StatusOK, ans)
	}
}
