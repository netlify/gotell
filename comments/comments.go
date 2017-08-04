package comments

import "regexp"

type RawComment struct {
	ID       string `json:"id"`
	ParentID string `json:"parent"`
	Author   string `json:"author"`
	Twitter  string `json:"twitter"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
	URL      string `json:"www"`
	IP       string `json:"ip"`
	Body     string `json:"body"`
	Date     string `json:"date"`
}

type ParsedComment struct {
	ID       string `json:"id"`
	ParentID string `json:"parent"`
	Author   string `json:"author"`
	Verified bool   `json:"verified"`
	Twitter  string `json:"twitter"`
	MD5      string `json:"md5"`
	URL      string `json:"www"`
	Body     string `json:"body"`
	Date     string `json:"date"`
}

var urlRegexp = regexp.MustCompile("(?i)https?://")

func (r *RawComment) IsSuspicious() bool {
	return urlRegexp.MatchString(r.Author + r.Twitter + r.Email + r.Body)
}
