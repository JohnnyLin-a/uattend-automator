package main

import (
	"log"
	"os"

	"github.com/johnnylin-a/uattend-automator/v2/pkg/api"
)

func main() {
	log.Println("Executing uAttend automator")

	log.Println("Init api...")
	err := api.InitApi()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}
