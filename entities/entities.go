package entities

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type ResultProcessor interface {
	Run(in <-chan Output)
}

type Parser interface {
	Parse(ctx context.Context, body []byte) (Individual, error)
}

type Job interface {
	GetID() string
	GetName() string
	GetUrl() string
	GetParser() Parser
}

type Output struct {
	Job        Job `json:"-"`
	Error      error
	StatusCode int
	Body       []byte `json:"-"`
	Individual Individual
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
	return fmt.Sprintf(`<url="%s" error="%v" status="%d" individual="%s">`,
		o.Job.GetUrl(), o.Error, o.StatusCode, o.Individual.String())

}

type Individual struct {
	Name               string
	RegistrationNumber string
	Status             string
	RegistrantType     string
	FirstRegisteredOn  string
	CurrentPeriodFrom  string
	CurrentPeriodUntil string
	Qualifications     []string
}

func (o Individual) String() string {
	return fmt.Sprintf(`<name="%s" registrationNumber="%s" Status="%s" RegistrantType="%s FirstRegisteredOn=%s CurrentPeriodFrom="%s" CurrentPeriodUntil="%s" Qualifications="%s">`,
		o.Name, o.RegistrationNumber, o.Status, o.RegistrantType, o.FirstRegisteredOn, o.CurrentPeriodFrom, o.CurrentPeriodUntil, strings.Join(o.Qualifications, ","))
}

type BaseJob struct {
	uuid   string
	parser Parser
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
	const baseUrl = "https://olr.gdc-uk.org/SearchRegister/SearchResult?RegistrationNumber="
	return baseUrl + url.QueryEscape(o.regNum)
}

func (o SearchRegistrationNumberJob) GetParser() Parser {
	return o.parser
}

func (o SearchRegistrationNumberJob) GetID() string {
	return o.uuid
}
