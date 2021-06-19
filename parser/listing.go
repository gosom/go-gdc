package parser

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/gosom/go-gdc/entities"
)

type ListingResultParser struct {
}

func NewListinResultParser() *ListingResultParser {
	ans := ListingResultParser{}
	return &ans
}

func (o *ListingResultParser) Parse(ctx context.Context, body []byte) ([]entities.Individual, string, error) {
	os.WriteFile("1.html", body, 0777)
	var ans []entities.Individual
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return ans, "", err
	}
	sel := "table#search-results tr > td:nth-of-type(3)"
	dom.Find(sel).Each(func(i int, s *goquery.Selection) {
		item := entities.Individual{
			RegistrationNumber: strings.TrimSpace(s.Text()),
		}
		ans = append(ans, item)
	})
	next, exists := dom.Find("li.PagedList-skipToNext>a").Attr("href")
	if !exists {
		next = ""
	}

	return ans, next, nil

}
