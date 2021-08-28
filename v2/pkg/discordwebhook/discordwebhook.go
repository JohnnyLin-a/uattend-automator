package discordwebhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Details struct {
	Webhook string
	Mention string
}

type discordConfig struct {
	Details Details `json:"Discord"`
}

var (
	config discordConfig
	isInit bool = false
)

func Init() error {
	if isInit {
		return errors.New("already init")
	}
	isInit = true
	err := loadConfig()
	if err != nil {
		return err
	}
	return nil
}

func loadConfig() error {
	log.Println("Loading discord config...")
	// Load config from os env UATTEND_CONFIG first
	jsonBytes := []byte(os.Getenv("UATTEND_CONFIG"))

	// Otherwise load config from config.json
	if len(jsonBytes) == 0 {
		jsonFile, err := os.Open("config.json")
		if err != nil {
			return errors.New("no config found")
		}
		defer jsonFile.Close()
		jsonBytes, err = ioutil.ReadAll(jsonFile)
		if err != nil {
			return errors.New("config file corrupted?")
		}
	}
	err := json.Unmarshal(jsonBytes, &config)
	if err != nil {
		return errors.New("config file json syntax error")
	}
	return nil
}

func Notify(count int, apierr error) error {
	log.Println("Notifying on Discord")
	body := make(map[string]interface{})
	if apierr != nil {
		body["content"] = "<@" + config.Details.Mention + "> Automated " + strconv.Itoa(count) + " row(s) with error: " + apierr.Error()
	} else {
		body["content"] = "<@" + config.Details.Mention + "> Automated " + strconv.Itoa(count) + " row(s)"
	}
	byteBody, _ := json.Marshal(body)
	responseBody := bytes.NewBuffer(byteBody)
	resp, err := http.Post(config.Details.Webhook, "application/json", responseBody)
	if err != nil {
		return err
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return errors.New("Network request was not 2xx: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
