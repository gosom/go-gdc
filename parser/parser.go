package parser

import (
	"bytes"
	"context"
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

func (o *SingleResultParser) Parse(ctx context.Context, body []byte) ([]entities.Individual, string, error) {
	var ans []entities.Individual
	item := entities.Individual{}
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return ans, "", err
	}

	details := dom.Find("#registrant-details>div.card")
	item.Name = strings.TrimSpace(details.Find("div.card-header>h2").Text())
	cardBody := details.Find("div.card-body")
	cardBody.Find("div.row").Each(func(i int, s *goquery.Selection) {
		if i > 0 && i < 11 {
			cols := s.Find("div")
			if cols.Length() == 3 {
				key := strings.TrimSpace(cols.Eq(0).Text())
				value := strings.TrimSpace(cols.Eq(1).Text())
				switch key {
				case "Dental Care Professional Titles:":
					item.ProfessionalTitles = value
				case "Registration Number:":
					item.RegistrationNumber = value
				case "Status:":
					item.Status = value
				case "Registrant Type:":
					item.RegistrantType = value
				case "First Registered on:":
					item.FirstRegisteredOn = value
				case "Current period of registration from:":
					sep := "until:"
					parts := strings.Split(value, sep)
					if len(parts) == 2 {
						item.CurrentPeriodFrom = strings.TrimSpace(parts[0])
						item.CurrentPeriodUntil = strings.TrimSpace(parts[1])
					} else {
						item.CurrentPeriodFrom = value
					}
				case "Qualifications:":
					lines := strings.Split(value, "\n")
					for i := range lines {
						item.Qualifications = append(item.Qualifications, strings.TrimSpace(lines[i]))
					}
				}
			}
		}
	})
	if len(item.RegistrationNumber) > 0 {
		ans = append(ans, item)
	}

	return ans, "", nil
}
