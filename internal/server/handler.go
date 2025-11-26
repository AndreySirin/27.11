package server

import (
	"encoding/json"
	"github.com/AndreySirin/newProject-28-11/internal/entity"
	"net/http"
)

func (s *Server) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	var request entity.RequestTask
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task := s.manager.RequestToService(request)

	err := s.manager.Db.SaveTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.manager.Ch <- task.ID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(task.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestTaskList
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pdfBytes, err := s.manager.GenerateReport(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(pdfBytes)
	if err != nil {
		s.lg.Error("failed to write pdf response: %v", err)
		return
	}
}
