package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/guregu/kami"
	"github.com/netlify/netlify-comments/comments"
	"github.com/netlify/netlify-comments/conf"
	"github.com/rs/cors"
	"github.com/zenazn/goji/web/mutil"
)

const defaultVersion = "unknown version"

type Server struct {
	handler  http.Handler
	config   *conf.Configuration
	client   *github.Client
	settings *settings
	mutex    sync.Mutex
	version  string
}

func (s *Server) postComment(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	entryPath := ctx.Value("path").(string)

	w.Header().Set("Content-Type", "application/json")

	settings := s.getSettings()
	for _, ip := range settings.BannedIPs {
		if req.RemoteAddr == ip {
			w.Header().Add("X-Banned", "IP-Banned")
			fmt.Fprintln(w, "{}")
			return
		}
	}

	entryData, err := s.entryData(entryPath)
	if err != nil {
		jsonError(w, fmt.Sprintf("Unable to read entry data: %v", err), 400)
		return
	}
	if settings.TimeLimit != 0 && time.Now().Sub(entryData.CreatedAt) > time.Duration(settings.TimeLimit) {
		jsonError(w, "Thread is closed for new comments", 401)
		return
	}

	comment := &comments.RawComment{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(comment); err != nil {
		jsonError(w, fmt.Sprintf("Error decoding JSON body: %v", err), 422)
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
	comment.ID = fmt.Sprintf("%v", time.Now().UnixNano())

	parts := strings.Split(s.config.API.Repository, "/")
	pathname := path.Join(
		s.config.Threads.Source,
		entryData.Thread,
		fmt.Sprintf("%v.json", (time.Now().UnixNano()/1000000)),
	)
	content, _ := json.Marshal(comment)
	message := "Add Comment"
	_, _, err = s.client.Repositories.CreateFile(parts[0], parts[1], pathname, &github.RepositoryContentFileOptions{
		Message: &message,
		Content: content,
	})

	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to write comment: %v", err), 500)
		return
	}

	parsedComment := comments.ParseRaw(comment)
	response, _ := json.Marshal(parsedComment)
	w.Write(response)
}

// Index endpoint
func (s *Server) index(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sendJSON(w, 200, map[string]string{
		"version":     s.version,
		"name":        "Netlify Comments",
		"description": "Netlify Comments is an API and build tool for handling large amounts of comments for JAMstack products",
	})
}

// ListenAndServe starts the Comments Server
func (s *Server) ListenAndServe() error {
	l := fmt.Sprintf("%v:%v", s.config.API.Host, s.config.API.Port)
	logrus.Infof("Netlify Comments API started on: %s", l)
	return http.ListenAndServe(l, s.handler)
}

func NewServer(config *conf.Configuration, githubClient *github.Client) *Server {
	return NewServerWithVersion(config, githubClient, defaultVersion)
}

func NewServerWithVersion(config *conf.Configuration, githubClient *github.Client, version string) *Server {
	s := &Server{
		config:  config,
		client:  githubClient,
		version: version,
	}

	mux := kami.New()
	mux.LogHandler = logHandler
	mux.Use("/", timeRequest)
	mux.Use("/", jsonTypeRequired)
	mux.Get("/", s.index)
	mux.Post("/*path", s.postComment)

	s.handler = cors.Default().Handler(mux)
	return s
}

func timeRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return context.WithValue(ctx, "_netlify_comments_timing", time.Now())
}

func logHandler(ctx context.Context, wp mutil.WriterProxy, req *http.Request) {
	start := ctx.Value("_netlify_comments_timing").(time.Time)
	logrus.WithFields(logrus.Fields{
		"method":   req.Method,
		"path":     req.URL.Path,
		"status":   wp.Status(),
		"duration": time.Since(start),
	}).Info("")
}

func jsonTypeRequired(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	if r.Method == "POST" && r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", 422)
		return nil
	}
	return ctx
}

func sendJSON(w http.ResponseWriter, status int, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.Encode(obj)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.Encode(map[string]string{"msg": message})
}
