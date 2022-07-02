package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

var _ http.Handler = &Server{}

type Server struct {
	router *chi.Mux
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServer(r *Resource) *Server {
	return &Server{
		router: r.Mux,
	}
}

type Resource struct {
	*chi.Mux
}

func (tr Resource) Routes() *Resource {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", tr.Index())
	tr.Mux = r
	return &tr
}

func NewResource() *Resource {
	return Resource{}.Routes()
}

func (rs Resource) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, `{ title:"Taskify" } `)

	}
}
