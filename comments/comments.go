package comments

type RawComment struct {
	ID       string `json:"id"`
	ParentID string `json:"parent"`
	Author   string `json:"author"`
	Twitter  string `json:"twitter"`
	Email    string `json:"email"`
	URL      string `json:"www"`
	IP       string `json:"ip"`
	Body     string `json:"body"`
	Date     string `json:"date"`
}

type ParsedComment struct {
	ID       string `json:"id"`
	ParentID string `json:"parent"`
	Author   string `json:"author"`
	Twitter  string `json:"twitter"`
	MD5      string `json:"md5"`
	URL      string `json:"www"`
	Body     string `json:"body"`
	Date     string `json:"date"`
}
