package main

import (
	"log"
	"os"

	"github.com/gosom/go-gdc/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}
