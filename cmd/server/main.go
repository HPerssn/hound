package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/hperssn/hound/internal/domain"
	httpapi "github.com/hperssn/hound/internal/http"
	"github.com/hperssn/hound/internal/runner"
)

func main() {
	manager := runner.NewSessionManager()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/sessions", startSession(manager))
	r.Get("/sessions/{id}", getSession(manager))
	r.Post("/sessions{id}/stop", stopSession(manager))
	r.Get("/sessions/{id}/events", httpapi.StreamSessionEvents(manager))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func startSession(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID        string `json:"id"`
			TargetSec int    `json:"targetSec"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s := domain.NewSession(req.ID, "", req.TargetSec)

		if err := m.StartSession(s); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(s)
	}
}

func getSession(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		s, ok := m.GetSession(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(s)
	}
}

func stopSession(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := m.StopSession(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
