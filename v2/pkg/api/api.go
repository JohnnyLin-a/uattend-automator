package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/johnnylin-a/uattend-automator/v2/pkg/generic"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
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
}

type apiConfig struct {
	Credentials credentials
	OrgURL      string
	SkipDates   []skipDate
	Workdays    []int
	Behavior    apiBehavior
}

var (
	debug             bool
	isInit            bool = false
	config            apiConfig
	validPunchTypes   = []string{"In/Out", "Break", "Lunch", "Benefit"}
	validBenefitTypes = []string{"VAC - Vacation", "SIC - Sick", "HOL - Holiday", "OTH - Other"}
)

func InitApi() error {
	if isInit {
		return errors.New("already init")
	}
	isInit = true
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
		switch v {
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
		if _, err := strconv.ParseFloat(config.Behavior.BenefitHours, 64); len(config.Behavior.BenefitHours) == 0 || err != nil {
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

func createWebdriver() (*selenium.Service, selenium.WebDriver, error) {
	// Init Selenium
	debugStr := os.Getenv("DEBUG")
	var err error
	debug, err = strconv.ParseBool(debugStr)
	if err != nil {
		debug = false
	}
	headless := !debug
	const (
		seleniumPath    = "deps/selenium-server-standalone-3.141.59.jar"
		geckoDriverPath = "/usr/bin/geckodriver"
		port            = 4444
	)
	selenium.SetDebug(debug)
	opts := []selenium.ServiceOption{}
	if debug {
		opts = append(opts, selenium.Output(os.Stderr))
	}

	service, err := selenium.NewGeckoDriverService(geckoDriverPath, port, opts...)
	if err != nil {
		return nil, nil, errors.New("cannot init selenium service")
	}

	caps := selenium.Capabilities{}
	if headless {
		caps.AddFirefox(firefox.Capabilities{
			Args: []string{"--headless"},
		})
	}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return nil, nil, errors.New("cannot connect to webdriver")
	}
	return service, wd, nil
}

func Execute() error {
	service, wd, err := createWebdriver()
	if err != nil {
		return err
	}
	defer service.Stop()
	if !debug {
		defer wd.Quit()
	}

	// Login
	log.Println("Login...")
	if err := wd.Get(config.OrgURL); err != nil {
		return errors.New("cannot go to org url")
	}
	usernameElem, err := wd.FindElement(selenium.ByCSSSelector, "#txtUserName")
	if err != nil {
		return errors.New("cannot find login username field")
	}
	passwordElem, err := wd.FindElement(selenium.ByCSSSelector, "#txtPassword")
	if err != nil {
		return errors.New("cannot find login password field")
	}
	loginBtnElem, err := wd.FindElement(selenium.ByCSSSelector, "#loginIn")
	if err != nil {
		return errors.New("cannot find login button element")
	}
	err = usernameElem.SendKeys(config.Credentials.Login)
	if err != nil {
		return errors.New("cannot type login username")
	}
	err = passwordElem.SendKeys(config.Credentials.Password)
	if err != nil {
		return errors.New("cannot type login password")
	}
	err = loginBtnElem.Click()
	if err != nil {
		return errors.New("cannot click login button")
	}
	// TODO: Check if login successfully

	// Wait for login to finish loading
	err = wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		_, err := wd.FindElement(selenium.ByCSSSelector, "#rowsInner>ul")
		if err != nil {
			return false, nil
		}
		return true, nil
	}, (time.Minute * 2))
	if err != nil {
		// No need to stop execution if wait failed
		log.Println("WARNING: Failed to wait after login, perhaps login failed?")
	}

	timesheetRowsListSelector := "#rowsInner>ul>li"
	timesheetRows, err := wd.FindElements(selenium.ByCSSSelector, timesheetRowsListSelector)
	if err != nil {
		return errors.New("cannot find individual timesheet rows")
	}

	// Loop through rows
	log.Println("Checking punch sheet...")
	for i, v := range timesheetRows {
		if i >= 14 {
			break
		}
		// Check if punch is already entered
		// Get delete button if it exists
		temps, err := wd.FindElements(selenium.ByCSSSelector, "a[class^='js-toggle-delete-punch'][title='Delete'][data-parent-index='"+strconv.Itoa(i)+"']")
		if err != nil {
			return errors.New("cannot find already punched rows")
		}
		if len(temps) > 0 {
			// Trash button found, punch already done
			log.Println("Punch already done for row", (i + 1))
			continue
		}

		// get Date
		temp, err := v.FindElement(selenium.ByCSSSelector, "ul>li>div>div")
		if err != nil {
			return errors.New("cannot find date for this row " + strconv.Itoa(i+1))
		}
		strRaw, err := temp.Text()
		if err != nil {
			return errors.New("cannot get row " + strconv.Itoa(i+1) + "'s date as text")
		}
		startStrDate := strings.LastIndex(strRaw, "\n")
		if startStrDate == -1 {
			return errors.New("row " + strconv.Itoa(i+1) + "'s date as text is not what it should be")
		}

		rowDate, err := time.Parse("01/02/06", strRaw[startStrDate+1:]) // mm/dd/yy -> 01/02/06
		if err != nil {
			return errors.New("cannot parse row " + strconv.Itoa(i+1) + "'s date format")
		}

		// Check for skip dates
		toSkip := false
		for _, v2 := range config.SkipDates {
			// Already validated parse
			skipDateStart, _ := time.Parse("2006-01-02", v2.Start)
			skipDateEnd, _ := time.Parse("2006-01-02", v2.End)
			if (rowDate.After(skipDateStart) && rowDate.Before(skipDateEnd)) || rowDate.Equal(skipDateStart) || rowDate.Equal(skipDateEnd) {
				toSkip = true
				break
			}
		}
		if toSkip {
			log.Println("Skipping row due to skipDate being set.")
			continue
		}

		// Check if workday
		if generic.InArray(int(rowDate.Weekday()), config.Workdays) < 0 {
			log.Println("skip row", (i + 1), "not a workday")
			continue
		}

		// Click add punch in time
		temp, err = wd.FindElement(selenium.ByCSSSelector, "a[data-date='"+rowDate.Format("2006-01-02")+"'][title='Add']")
		if err != nil {
			return errors.New("cannot find add timeslot button")
		}
		err = temp.Click()
		if err != nil {
			return errors.New("click add timeslot button failed")
		}

		// Wait for modal to be loaded
		wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
			l, err := wd.FindElements(selenium.ByCSSSelector, "#modalContainer")
			if err != nil {
				return false, err
			}
			if len(l) > 0 {
				return true, nil
			}
			return false, nil
		}, time.Minute*1)

		// Click punch type dropdown
		temp, err = wd.FindElement(selenium.ByCSSSelector, "#addPunchForm>div.col-group>div>.select-wrapper>input.select-dropdown.dropdown-trigger")
		if err != nil {
			return errors.New("cannot find punch type dropdown")
		}
		err = temp.Click()
		if err != nil {
			return errors.New("cannot expand punch type dropdown")
		}
		temps, err = wd.FindElements(selenium.ByCSSSelector, "#addPunchForm>div.col-group>div>.select-wrapper>ul>li")
		if err != nil || len(temps) == 0 {
			return errors.New("cannot find punch type dropdown elements")
		}
		tempMap := make(map[string]selenium.WebElement)
		for i, v := range temps {
			s, err := v.GetAttribute("class")
			if err == nil && s == "disabled" {
				continue
			}
			s, err = v.Text()
			if err != nil {
				return errors.New("cannot find punch type #" + strconv.Itoa(i+1))
			}
			tempMap[s] = v
		}

		if _, ok := tempMap[config.Behavior.PunchType]; !ok {
			return errors.New("punch type \"" + config.Behavior.PunchType + "\" does not exist")
		}
		temp = tempMap[config.Behavior.PunchType]
		tempMap = nil
		temp.Click()

		if config.Behavior.PunchType == "Benefit" {
			// Wait for benefit types to load
			wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
				temps, err := wd.FindElements(selenium.ByCSSSelector, "#benefitWrapper>div>div>div>ul>li")
				if err != nil || len(temps) == 0 {
					return false, err
				}
				return true, nil
			}, time.Minute * 1)

			// Click benefit type dropdown
			temp, err = wd.FindElement(selenium.ByCSSSelector, "#benefitWrapper>div>div>div>input.select-dropdown.dropdown-trigger")
			if err != nil {
				return errors.New("cannot find benefit type dropdown")
			}
			err = temp.Click()
			if err != nil {
				return errors.New("cannot expand benefit type dropdown")
			}
			// Get list of benefit types
			temps, err = wd.FindElements(selenium.ByCSSSelector, "#benefitWrapper>div>div>div>ul>li>span")
			if err != nil || len(temps) == 0 {
				return errors.New("cannot find benefit type list")
			}
			tempMap = make(map[string]selenium.WebElement)
			for _, v := range temps {
				s, err := v.GetAttribute("class")
				if err == nil && s == "disabled" {
					continue
				}
				s, err = v.Text()
				if err != nil {
					return errors.New("cannot find benefit type #" + strconv.Itoa(i+1))
				}
				tempMap[s] = v
			}
			if _, ok := tempMap[config.Behavior.BenefitType]; !ok {
				// TO REMOVE AFTER 
				for k := range tempMap {
					log.Println(k)
				}
				return errors.New("benefit type \"" + config.Behavior.BenefitType + "\" does not exist")
			}
			temp = tempMap[config.Behavior.BenefitType]
			tempMap = nil
			err = temp.Click()
			if err != nil {
				return errors.New("cannot click benefit type " + config.Behavior.BenefitType)
			}
			
			// Set Benefit Hours
			temp, err = wd.FindElement(selenium.ByCSSSelector, "#benefit_hours")
			if err != nil {
				return errors.New("cannot find benefit hours input field")
			}
			err = temp.Click()
			if err != nil {
				return errors.New("cannot click benefit hours input field")
			}
			err = temp.SendKeys(config.Behavior.BenefitHours)
			if err != nil {
				return errors.New("cannot type in benefit hours")
			}

			// Click save and close
			_, err = wd.FindElement(selenium.ByCSSSelector, "#addPunch>div>div>div>span>button[class^='js-add-confirm'][data-next='false']")
			if err != nil{
				return errors.New("cannot find submit and close button")
			}
			log.Println("stopped at found submit and close button")
		} else {
			// TODO: To handle In/Out, Break, Lunch
			log.Println("WARNING: did not implement In/Out, Break, Lunch")
		}
		return nil
	}

	// Get this week's rows html elements
	/*
		Loop through rows
			Click +
			Set Punch Type
			if Benefit
				Set BenefitType
				Set BenefitHours
			else
				Set InTime
				Set OutTime
			Click "Save and Close" and wait for model to be gone
		Notify result to user through discord
	*/
	return nil
}
