package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"

	"mail2calendar/internal/domain/authentication"
	"mail2calendar/internal/domain/health"
	"mail2calendar/internal/middleware"
	"mail2calendar/internal/utility/respond"
)

func (s *Server) InitDomains() {
	s.initVersion()
	s.initSwagger()
	s.initAuthentication()
	s.initHealth()
}

func (s *Server) initVersion() {
	s.router.Route("/version", func(router chi.Router) {
		router.Use(middleware.Json)

		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			respond.Json(w, http.StatusOK, map[string]string{"version": s.Version})
		})
	})
}

func (s *Server) initHealth() {
	newHealthRepo := health.NewRepo(s.sqlx)
	newHealthUseCase := health.New(newHealthRepo)
	health.RegisterHTTPEndPoints(s.router, newHealthUseCase)
}

//go:embed docs/*
var swaggerDocsAssetPath embed.FS

func (s *Server) initSwagger() {
	if s.Config().Api.RunSwagger {
		docsPath, err := fs.Sub(swaggerDocsAssetPath, "docs")
		if err != nil {
			panic(err)
		}

		fileServer := http.FileServer(http.FS(docsPath))

		s.router.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
		})
		s.router.Handle("/swagger/", http.StripPrefix("/swagger", middleware.ContentType(fileServer)))
		s.router.Handle("/swagger/*", http.StripPrefix("/swagger", middleware.ContentType(fileServer)))
	}
}

func (s *Server) initAuthentication() {
	repo := authentication.NewRepo(s.db, s.session)
	authentication.RegisterHTTPEndPoints(s.router, s.session, repo)
}
