package parser

import (
	"bytes"
	"context"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/gosom/go-gdc/entities"
)

type SingleResultParser struct {
}

func NewSingleResultParser() *SingleResultParser {
	ans := SingleResultParser{}
	return &ans
}

func (o *SingleResultParser) Parse(ctx context.Context, body []byte) (entities.Individual, error) {
	ans := entities.Individual{}
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		log.Println(err.Error())
		return ans, err
	}

	details := dom.Find("#registrant-details>div.card")
	ans.Name = strings.TrimSpace(details.Find("div.card-header>h2").Text())
	cardBody := details.Find("div.card-body")
	cardBody.Find("div.row").Each(func(i int, s *goquery.Selection) {
		if i > 0 && i < 7 {
			cols := s.Find("div")
			if cols.Length() == 3 {
				key := strings.TrimSpace(cols.Eq(0).Text())
				value := strings.TrimSpace(cols.Eq(1).Text())
				switch key {
				case "Registration Number:":
					ans.RegistrationNumber = value
				case "Status:":
					ans.Status = value
				case "Registrant Type:":
					ans.RegistrantType = value
				case "First Registered on:":
					ans.FirstRegisteredOn = value
				case "Current period of registration from:":
					sep := "until:"
					parts := strings.Split(value, sep)
					if len(parts) == 2 {
						ans.CurrentPeriodFrom = strings.TrimSpace(parts[0])
						ans.CurrentPeriodUntil = strings.TrimSpace(parts[1])
					} else {
						ans.CurrentPeriodFrom = value
					}
				case "Qualifications:":
					lines := strings.Split(value, "\n")
					for i := range lines {
						ans.Qualifications = append(ans.Qualifications, strings.TrimSpace(lines[i]))
					}
				}
			}
		}
	})

	return ans, nil
}
