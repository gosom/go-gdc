package finisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gosom/go-gdc/entities"
)

type SimpleStdout struct {
}

func (o *SimpleStdout) Run(in <-chan entities.Output) {
	for el := range in {
		fmt.Println(el)
	}
}

type SimpleJson struct {
}

func (o *SimpleJson) Run(in <-chan entities.Output) {
	for el := range in {
		b, err := json.Marshal(&el)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
}

type IndividualRepo interface {
	BulkInsert(ctx context.Context, items []entities.Individual) error
}

type DbSaver struct {
	repo IndividualRepo
}

func NewDbSaver(repo IndividualRepo) *DbSaver {
	ans := DbSaver{
		repo: repo,
	}
	return &ans
}

func (o *DbSaver) Run(in <-chan entities.Output) {
	ctx := context.Background()
	for el := range in {
		fmt.Println(el)
		if len(el.Individuals) > 0 {
			if err := o.repo.BulkInsert(ctx, el.Individuals); err != nil {
				panic(err)
			}
		}
	}
}
