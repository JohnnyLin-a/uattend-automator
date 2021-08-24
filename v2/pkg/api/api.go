package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/johnnylin-a/uattend-automator/v2/pkg/generic"
)

type credentials struct {
	Login    string
	Password string
}

type skipDate struct {
	Start string
	End   string
}

type apiBehavior struct {
	PunchType    string
	InTime       string
	OutTime      string
	BenefitType  string
	BenefitHours string
	Notes        string
}

type apiConfig struct {
	Credentials credentials
	OrgURL      string
	SkipDates   []skipDate
	Workdays    []*int
	Behavior    apiBehavior
}

var (
	config            apiConfig
	validPunchTypes   = []string{"In/Out", "Break", "Lunch", "Benefit"}
	validBenefitTypes = []string{"VAC - Vacation", "SIC - Sick", "HOL - Holiday", "OTH - Other"}
)

func InitApi() error {
	err := loadConfig()
	if err != nil {
		return err
	}
	err = validateConfig()
	if err != nil {
		return err
	}
	return nil
}

func validateConfig() error {
	// Check credentials
	if len(config.Credentials.Login) == 0 || len(config.Credentials.Password) == 0 {
		return errors.New("credentials not set up in config")
	}

	// Check OrgURL
	if _, err := url.ParseRequestURI(config.OrgURL); err != nil {
		return errors.New("invalid orgurl")
	}
	if u, err := url.Parse(config.OrgURL); err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("invalid orgurl")
	}

	// Check SkipDates
	if len(config.SkipDates) > 0 {
		for i, v := range config.SkipDates {
			if len(v.Start) == 0 {
				return errors.New("Start date for Date #" + strconv.Itoa(i+1) + " invalid. Did you typo 'start' instead of 'Start'?")
			}
			if len(v.End) == 0 {
				return errors.New("End date for Date #" + strconv.Itoa(i+1) + " invalid. Did you typo 'end' instead of 'End'?")
			}
			_, err := time.Parse("2006-01-02", v.Start)
			if err != nil {
				return errors.New("Start Date #" + strconv.Itoa(i+1) + " invalid format, must be YYYY-MM-DD")
			}
			_, err = time.Parse("2006-01-02", v.End)
			if err != nil {
				return errors.New("end Date #" + strconv.Itoa(i+1) + " invalid format, must be YYYY-MM-DD")
			}
		}
	}

	// Check Workdays
	for i, v := range config.Workdays {
		if v == nil {
			return errors.New("workdays #" + strconv.Itoa(i+1) + " is not a valid weekday, must be numbers between 0-6")
		}
		switch *v {
		case int(time.Sunday):
			fallthrough
		case int(time.Monday):
			fallthrough
		case int(time.Tuesday):
			fallthrough
		case int(time.Wednesday):
			fallthrough
		case int(time.Thursday):
			fallthrough
		case int(time.Friday):
			fallthrough
		case int(time.Saturday):
			continue
		default:
			return errors.New("workdays #" + strconv.Itoa(i+1) + " is not a valid weekday, must be numbers between 0-6")
		}
	}

	// Validate Behavior
	if i := generic.InArray(config.Behavior.PunchType, validPunchTypes); i < 0 {
		return errors.New("invalid punchtype, must be value of either: " + strings.Join(validPunchTypes, ", "))
	}
	if config.Behavior.PunchType == "Benefit" {
		if i := generic.InArray(config.Behavior.BenefitType, validBenefitTypes); i < 0 {
			return errors.New("invalid benefittype, must be value of either: " + strings.Join(validBenefitTypes, ", "))
		}
		if _, err := strconv.ParseFloat(config.Behavior.BenefitHours, 64); len(config.Behavior.BenefitHours) == 0 || err != nil{
			return errors.New("benefit hours invalid")
		}
	} else {
		if len(config.Behavior.InTime) == 0 {
			return errors.New("InTime invalid")
		}
		if len(config.Behavior.OutTime) == 0 {
			return errors.New("OutTime invalid")
		}
		_, err := time.Parse("2006-01-02", config.Behavior.InTime)
		if err != nil {
			return errors.New("InTime invalid format, must be YYYY-MM-DD")
		}
		_, err = time.Parse("2006-01-02", config.Behavior.OutTime)
		if err != nil {
			return errors.New("outtime invalid format, must be YYYY-MM-DD")
		}

		if len(config.Behavior.InTime) == 0 {
			return errors.New("InTime invalid. Did you typo 'InTime' instead of 'InTime'?")
		}
		if len(config.Behavior.OutTime) == 0 {
			return errors.New("OutTime invalid. Did you typo 'outtime' instead of 'OutTime'?")
		}
		_, err = time.Parse("2006-01-02", config.Behavior.InTime)
		if err != nil {
			return errors.New("InTime invalid format, must be YYYY-MM-DD")
		}
		_, err = time.Parse("2006-01-02", config.Behavior.OutTime)
		if err != nil {
			return errors.New("OutTime invalid format, must be YYYY-MM-DD")
		}
	}
	return nil
}

func loadConfig() error {
	// Load config from os env UATTEND_CONFIG first
	jsonBytes := []byte(os.Getenv("UATTEND_CONFIG"))

	// Otherwise load config from config.json
	if len(jsonBytes) == 0 {
		jsonFile, err := os.Open("config.json")
		if err != nil {
			return errors.New("no config found")
		}
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
