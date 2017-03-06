package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var cleanPathRE = regexp.MustCompile("(^-+|-+$)")

type entryData struct {
	Thread    string
	CreatedAt time.Time
}

func (s *Server) entryData(path string) (*entryData, error) {
	resp, err := http.Get(s.config.API.SiteURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	metaTag := doc.Find("#gotell").First()
	if metaTag.Length() == 0 {
		return nil, fmt.Errorf("No script tag with id gotell found for '%v'", path)
	}
	entryData := &entryData{}
	if err := json.Unmarshal([]byte(metaTag.Text()), entryData); err != nil {
		return nil, err
	}

	if entryData.Thread == "" {
		entryData.Thread = strings.Replace(path, "/", "-", -1)
		entryData.Thread = cleanPathRE.ReplaceAllLiteralString(entryData.Thread, "")
	}

	return entryData, nil
}
