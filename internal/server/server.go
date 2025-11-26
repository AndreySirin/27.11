package server

import (
	"context"
	"errors"
	"github.com/AndreySirin/newProject-28-11/internal/taskManager"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	HttpServer *http.Server
	manager    *taskManager.TaskManager
	lg         *slog.Logger
}

func New(addr string, manager *taskManager.TaskManager, logger *slog.Logger) *Server {
	s := &Server{
		lg:      logger,
		manager: manager,
	}

	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/tasks", s.HandleCreateTask)
			r.Post("/report", s.HandleGetReport)
		})
	})
	s.HttpServer = &http.Server{
		Addr:    addr,
		Handler: r,
	}
	return s
}
func (s *Server) Run() {
	s.lg.Info("server is running")
	err := s.HttpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		s.lg.Info("server is stopped")
		return
	}
}

func (s *Server) ShutDown() {
	s.lg.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.HttpServer.Shutdown(ctx)
	if err != nil {
		s.lg.Error("error when stopping the server")
		return
	}
}
