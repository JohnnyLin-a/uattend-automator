package api

type credentials struct {
	Login    string
	Passowrd string
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
	Workdays    []int
	Behavior    apiBehavior
}

var (
	config apiConfig
)

func GetApi() {

}
