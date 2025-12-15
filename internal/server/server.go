// internal/server/server.go
package server

import (
	"maven-proxy/pkg/auth"
	"maven-proxy/pkg/config"
	"maven-proxy/pkg/repository"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config        *config.Config
	engine        *gin.Engine
	repositories  map[string]repository.Repository
	authenticator auth.Authenticator
}

func NewServer(cfg *config.Config, authenticator auth.Authenticator) *Server {
	s := &Server{
		config:        cfg,
		engine:        gin.Default(),
		repositories:  make(map[string]repository.Repository),
		authenticator: authenticator,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// GET 和 HEAD 不需要认证
	s.engine.GET("/:context/:repoId/*path", s.handleGet)
	s.engine.HEAD("/:context/:repoId/*path", s.handleGet)

	// PUT 需要认证
	s.engine.PUT("/:context/:repoId/*path",
		auth.Middleware(s.authenticator),
		s.handlePut)
}

func (s *Server) RegisterRepository(id string, repo repository.Repository) {
	s.repositories[id] = repo
}

func (s *Server) Run() error {
	addr := s.config.Listen + ":" + s.config.Port
	return s.engine.Run(addr)
}
