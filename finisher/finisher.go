package finisher

import (
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
