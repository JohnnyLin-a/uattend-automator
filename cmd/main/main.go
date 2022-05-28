package main

import (
	"log"
	"os"

	"github.com/johnnylin-a/uattend-automator/pkg/api"
	"github.com/johnnylin-a/uattend-automator/pkg/discordwebhook"
)

func main() {
	log.Println("Executing uAttend automator")

	log.Println("Init api...")
	apierr := api.InitApi()
	if apierr != nil {
		log.Println(apierr.Error())
		os.Exit(1)
	}
	apierr = api.Execute()
	if apierr != nil {
		log.Println(apierr.Error())
	}

	discorderr := discordwebhook.Init()
	if discorderr != nil {
		log.Println(discorderr.Error())
	} else {
		discorderr = discordwebhook.Notify(api.GetAutomatedRowsCount(), apierr)
		if discorderr != nil {
			log.Println(discorderr.Error())
		}
	}

	if apierr != nil {
		os.Exit(1)
	}
}
