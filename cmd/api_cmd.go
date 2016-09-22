package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-github/github"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const SettingsRefreshTime = 10 * time.Minute

var cleanPathRE = regexp.MustCompile("(^-+|-+$)")

func APICommand() *cobra.Command {
	APICmd := cobra.Command{
		Use:   "api",
		Short: "api",
		Run:   ServeAPI,
	}

	APICmd.Flags().IntP("port", "p", 9911, "the port to listen on")
	APICmd.Flags().StringP("site", "s", "", "URL of the website")
	APICmd.Flags().StringP("repo", "r", "", "user/repo for the GitHub repo")
	APICmd.Flags().StringP("token", "t", "", "GitHub access token")

	return &APICmd
}

type API struct {
	Site     string
	Repo     string
	Token    string
	Client   *github.Client
	settings *Settings
	mutex    sync.Mutex
}

type Settings struct {
	BannedIPs      []string `json:"banned_ips"`
	BannedKeywords []string `json:"banned_keywords"`
	BannedEmails   []string `json:"banned_emails"`
	TimeLimit      int      `json:"timelimit"`
	lastLoad       time.Time
}

type EntryData struct {
	Thread    string
	CreatedAt time.Time
}

func (s *Settings) Fresh() bool {
	return time.Now().Sub(s.lastLoad) < SettingsRefreshTime
}

func (a *API) Settings() *Settings {
	if a.settings != nil && a.settings.Fresh() {
		return a.settings
	}
	resp, err := http.Get(a.Site + "/gocomments/settings.json")
	if err != nil {
		return &Settings{}
	}
	defer resp.Body.Close()
	settings := &Settings{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(settings); err != nil {
		log.Printf("Error decoding settings: %v", err)
		return &Settings{}
	}
	settings.lastLoad = time.Now()
	a.mutex.Lock()
	a.settings = settings
	a.mutex.Unlock()
	return settings
}

func (a *API) EntryData(path string) (*EntryData, error) {
	resp, err := http.Get(a.Site + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	metaTag := doc.Find("#gocomments").First()
	if metaTag.Length() == 0 {
		return nil, fmt.Errorf("No script tag with id gocomments found for '%v'", path)
	}
	entryData := &EntryData{}
	if err := json.Unmarshal([]byte(metaTag.Text()), entryData); err != nil {
		return nil, err
	}

	if entryData.Thread == "" {
		entryData.Thread = strings.Replace(path, "/", "-", -1)
		entryData.Thread = cleanPathRE.ReplaceAllLiteralString(entryData.Thread, "")
	}

	return entryData, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		log.Printf("Warn - non POST request %v", req.URL)
		http.Error(w, "Only POST requests allowed", 405)
		return
	}
	if req.Header.Get("Content-Type") != "application/json" {
		log.Printf("Warn - non JSON request %v", req.URL)
		http.Error(w, "Content-Type must be application/json", 422)
		return
	}
	log.Printf("POST %v (%v)", req.URL, req.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	settings := a.Settings()
	for _, ip := range settings.BannedIPs {
		if req.RemoteAddr == ip {
			w.Header().Add("X-Banned", "IP-Banned")
			fmt.Fprintln(w, "{}")
			return
		}
	}

	entryData, err := a.EntryData(req.URL.Path)
	if err != nil {
		a.JsonError(w, fmt.Sprintf("Unable to read entry data: %v", err), 500)
		return
	}
	if settings.TimeLimit != 0 && time.Now().Sub(entryData.CreatedAt) > time.Duration(settings.TimeLimit) {
		a.JsonError(w, "Thread is closed for new comments", 401)
		return
	}

	comment := &RawComment{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(comment); err != nil {
		a.JsonError(w, fmt.Sprintf("Error decoding JSON body: %v", err), 422)
		return
	}

	for _, email := range settings.BannedEmails {
		if strings.Contains(comment.Email, email) || strings.Contains(comment.Body, email) || strings.Contains(comment.URL, email) {
			w.Header().Add("X-Banned", "Email-Banned")
			fmt.Fprintln(w, "{}")
			return
		}
	}

	for _, keyword := range settings.BannedKeywords {
		if strings.Contains(comment.Email, keyword) || strings.Contains(comment.Body, keyword) || strings.Contains(comment.URL, keyword) {
			w.Header().Add("X-Banned", "Keyword-Banned")
			fmt.Fprintln(w, "{}")
			return
		}
	}

	comment.IP = req.RemoteAddr
	comment.Date = time.Now().String()

	parts := strings.Split(a.Repo, "/")
	pathname := path.Join(
		"threads",
		entryData.Thread,
		fmt.Sprintf("%v.json", (time.Now().UnixNano()/1000000)),
	)
	content, _ := json.Marshal(comment)
	message := "Add Comment"
	_, _, err = a.Client.Repositories.CreateFile(parts[0], parts[1], pathname, &github.RepositoryContentFileOptions{
		Message: &message,
		Content: content,
	})

	if err != nil {
		a.JsonError(w, fmt.Sprintf("Failed to write comment: %v", err), 500)
		return
	}

	parsedComment := ParseComment(comment)
	response, _ := json.Marshal(parsedComment)
	w.Write(response)
}

func (a *API) JsonError(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.Encode(map[string]string{"msg": message})
}

func verifySite(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Expected 200 status code for %v, got %v", url, resp.StatusCode)
	}
	return nil
}

func verifyRepoAndToken(api *API) error {
	parts := strings.Split(api.Repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("Repo format must be owner/repo - was %v", api.Repo)
	}
	_, _, err := api.Client.Repositories.Get(parts[0], parts[1])
	return err
}

func ServeAPI(cmd *cobra.Command, args []string) {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		log.Fatalf("Failed to read command flags")
	}

	viper.SetEnvPrefix("COMMENT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	api := &API{
		Site:  viper.GetString("site"),
		Repo:  viper.GetString("repo"),
		Token: viper.GetString("token"),
	}

	if api.Site == "" || api.Repo == "" || api.Token == "" {
		log.Fatal("api requires --site --repo and --token")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: api.Token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	api.Client = github.NewClient(tc)

	if err := verifySite(api.Site); err != nil {
		log.Fatalf("Error verifying site: %v", err)
	}
	if err := verifyRepoAndToken(api); err != nil {
		log.Fatalf("Error verifying repo: %v", err)
	}

	port := viper.GetInt("port")
	log.Printf("Start API for %v on port %v (pushing comments to %v)", api.Site, port, api.Repo)
	panic(http.ListenAndServe(fmt.Sprintf(":%v", port), cors.Default().Handler(api)))
}
