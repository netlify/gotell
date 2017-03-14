package comments

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/guregu/kami"
	"github.com/netlify/gotell/conf"
)

type Server struct {
	handler http.Handler
	config  *conf.Configuration
}

// ListenAndServe starts the Comments Server
func (s *Server) ListenAndServe() error {
	l := fmt.Sprintf("%v:%v", s.config.Threads.Host, s.config.Threads.Port)
	logrus.Infof("GoTell Server started on: %s", l)
	return http.ListenAndServe(l, s.handler)
}

func (s *Server) serveFile(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	path := ctx.Value("path").(string)

	fs := filepath.Join(s.config.Threads.Destination, path)
	http.ServeFile(w, r, fs)
}

func NewServer(config *conf.Configuration) *Server {
	s := &Server{
		config: config,
	}

	mux := kami.New()
	mux.Get("/*path", s.serveFile)
	s.handler = mux
	return s
}
