package main

import (
	"log"

	"github.com/johnnylin-a/uattend-automator/v2/pkg/api"
)

func main() {
	log.Println("Executing uAttend automator")

	log.Println("Init api...")
	api.GetApi()
}
