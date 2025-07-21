package router

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/talx-hub/malerter/internal/api/middlewares"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
)

type Router struct {
	router *chi.Mux
	log    *logger.ZeroLogger
	secret string
}

func New(log *logger.ZeroLogger, secret string) *Router {
	return &Router{
		router: chi.NewRouter(),
		log:    log,
		secret: secret,
	}
}

type Handler interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	GetMetric(w http.ResponseWriter, r *http.Request)
	GetMetricJSON(w http.ResponseWriter, r *http.Request)
	DumpMetric(w http.ResponseWriter, r *http.Request)
	DumpMetricJSON(w http.ResponseWriter, r *http.Request)
	DumpMetricList(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
}

func (r *Router) SetRouter(h Handler) {
	r.router.Use(middlewares.Logging(r.log))

	r.router.Route("/", func(c chi.Router) {
		c.
			With(middlewares.WriteSignature(r.secret)).
			With(middlewares.Compress(r.log)).
			Get("/", h.GetAll)

		c.Route("/value", func(c chi.Router) {
			c.
				With(middleware.AllowContentType(constants.ContentTypeJSON)).
				With(middlewares.WriteSignature(r.secret)).
				With(middlewares.Decompress(r.log)).
				With(middlewares.Compress(r.log)).
				Post("/", h.GetMetricJSON)
			c.Get("/{type}/{name}", h.GetMetric)
		})

		c.Route("/update", func(c chi.Router) {
			c.
				With(middleware.AllowContentType(constants.ContentTypeJSON)).
				With(middlewares.WriteSignature(r.secret)).
				With(middlewares.Decompress(r.log)).
				With(middlewares.Compress(r.log)).
				Post("/", h.DumpMetricJSON)
			c.Post("/{type}/{name}/{val}", h.DumpMetric)
		})

		c.Route("/ping", func(c chi.Router) {
			c.Get("/", h.Ping)
		})

		c.Route("/updates", func(c chi.Router) {
			c.
				With(middleware.AllowContentType(constants.ContentTypeJSON)).
				With(middlewares.CheckSignature(r.secret)).
				With(middlewares.Decompress(r.log)).
				With(middlewares.Compress(r.log)).
				Post("/", h.DumpMetricList)
		})

		c.Route("/debug/pprof", func(c chi.Router) {
			c.HandleFunc("/", pprof.Index)
			c.HandleFunc("/cmdline", pprof.Cmdline)
			c.HandleFunc("/profile", pprof.Profile)
			c.HandleFunc("/symbol", pprof.Symbol)
			c.HandleFunc("/trace", pprof.Trace)

			for _, p := range []string{
				"allocs", "block", "goroutine", "heap", "mutex", "threadcreate",
			} {
				c.HandleFunc("/"+p, pprof.Handler(p).ServeHTTP)
			}
		})
	})

	r.router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	})
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
