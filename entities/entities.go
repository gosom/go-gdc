package entities

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

const (
	gdc = "https://olr.gdc-uk.org/SearchRegister/SearchResult"
)

type ResultProcessor interface {
	Run(in <-chan Output)
}

type Parser interface {
	Parse(ctx context.Context, body []byte) ([]Individual, string, error)
}

type Job interface {
	GetID() string
	GetName() string
	GetUrl() string
	GetParser() Parser
	GetMethod() string
	GetFormData() url.Values
}

type Output struct {
	Job         Job `json:"-"`
	Error       error
	StatusCode  int
	Body        []byte `json:"-"`
	Individuals []Individual
	Next        string `json:"-"`
}

func (o Output) MarshalJSON() ([]byte, error) {
	type Alias Output
	return json.Marshal(&struct {
		Url string
		Alias
	}{
		Url:   o.Job.GetUrl(),
		Alias: Alias(o),
	})
}

func (o Output) String() string {
	return fmt.Sprintf(`<url="%s" error="%v" status="%d" individualNum="%d">`,
		o.Job.GetUrl(), o.Error, o.StatusCode, len(o.Individuals))

}

type Individual struct {
	FirstName          string
	LastName           string
	RegistrationNumber string
	Status             string
	RegistrantType     string
	ProfessionalTitles string
	Specialty          string
	FirstRegisteredOn  string
	CurrentPeriodFrom  string
	CurrentPeriodUntil string
	Qualifications     []string
}

func (o Individual) String() string {
	return fmt.Sprintf(`<name="%s" registrationNumber="%s" Status="%s" RegistrantType="%s FirstRegisteredOn=%s CurrentPeriodFrom="%s" CurrentPeriodUntil="%s" Qualifications="%s">`,
		o.LastName, o.RegistrationNumber, o.Status, o.RegistrantType, o.FirstRegisteredOn, o.CurrentPeriodFrom, o.CurrentPeriodUntil, strings.Join(o.Qualifications, ","))
}

type BaseJob struct {
	uuid   string
	parser Parser
}

type DiscoverJob struct {
	BaseJob
	postCode string
	formData map[string]string
	method   string
	page     string
}

func NewDiscoverJob(p Parser, postCode string, method string, page string) DiscoverJob {
	ans := DiscoverJob{
		postCode: postCode,
		formData: map[string]string{
			"olRegister":               "all",
			"FirstNameSoundsLike":      "false",
			"SurnameSoundsLike":        "false",
			"IncludeErasedRegistrants": "false",
			"SortAscending":            "true",
		},
		method: method,
		page:   page,
	}

	ans.parser = p
	ans.uuid = uuid.New().String()
	return ans
}

func (o DiscoverJob) GetFormData() url.Values {
	form := url.Values{}
	for k, v := range o.formData {
		form.Add(k, v)
	}
	form.Add("Postcode", o.postCode)
	if o.page != "" {
		form.Add("page", o.page)
	}
	return form
}

func (o DiscoverJob) GetMethod() string {
	return o.method
}

func (o DiscoverJob) GetName() string {
	return "DiscoverJob"
}

func (o DiscoverJob) GetUrl() string {
	if o.method == "GET" {
		var params []string
		for k, v := range o.formData {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		params = append(params, fmt.Sprintf("Postcode=%s", url.QueryEscape(o.postCode)))
		params = append(params, fmt.Sprintf("page=%s", o.page))
		return gdc + "s?" + strings.Join(params, "&")
	} else {
		return gdc + "s"
	}
}

func (o DiscoverJob) GetParser() Parser {
	return o.parser
}

func (o DiscoverJob) GetID() string {
	return o.uuid
}

type SearchRegistrationNumberJob struct {
	BaseJob
	regNum string
}

func NewSearchRegistrationNumberJob(p Parser, regNum string) SearchRegistrationNumberJob {
	ans := SearchRegistrationNumberJob{
		regNum: regNum,
	}
	ans.parser = p
	ans.uuid = uuid.New().String()
	return ans
}

func (o SearchRegistrationNumberJob) GetName() string {
	return "SearchRegistrationNumberJob"
}

func (o SearchRegistrationNumberJob) GetUrl() string {
	return gdc + "?RegistrationNumber=" + url.QueryEscape(o.regNum)
}

func (o SearchRegistrationNumberJob) GetParser() Parser {
	return o.parser
}

func (o SearchRegistrationNumberJob) GetID() string {
	return o.uuid
}

func (o SearchRegistrationNumberJob) GetMethod() string {
	return "GET"
}

func (o SearchRegistrationNumberJob) GetFormData() url.Values {
	return nil
}
