package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

const settingsRefreshTime = 10 * time.Minute

type settings struct {
	BannedIPs      []string `json:"banned_ips"`
	BannedKeywords []string `json:"banned_keywords"`
	BannedEmails   []string `json:"banned_emails"`
	TimeLimit      int      `json:"timelimit"`
	lastLoad       time.Time
}

func (s *settings) fresh() bool {
	return time.Since(s.lastLoad) < settingsRefreshTime
}

func (s *Server) getSettings() *settings {
	if s.settings != nil && s.settings.fresh() {
		return s.settings
	}

	resp, err := http.Get(s.config.API.SiteURL + "/netlify-comments/settings.json")
	if err != nil {
		return &settings{}
	}

	defer resp.Body.Close()
	st := &settings{}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(st); err != nil {
		logrus.Warnf("Error decoding settings: %v", err)
		return st
	}

	st.lastLoad = time.Now()
	s.mutex.Lock()
	s.settings = st
	s.mutex.Unlock()

	return st
}
