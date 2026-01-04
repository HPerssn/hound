package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/hperssn/hound/internal/domain"
	"github.com/hperssn/hound/internal/http"
	"github.com/hperssn/hound/internal/runner"
)

func main() {
	manager := runner.NewSessionManager()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/sessions", startSession(manager))
	r.Get("/sessions/{id}", getSession(manager))
	r.Post("/sessions/{id}/steps/{idx}/start", startStep(manager))
	r.Post("/sessions/{id}/steps/{idx}/stop", stopStep(manager))
	r.Post("/sessions/{id}/stop", stopSession(manager))
	r.Get("/sessions/{id}/events", httpapi.StreamSessionEvents(manager))
	r.Get("/sessions/{id}/status", getSessionStatus(manager))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func startSession(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TargetSec int `json:"targetSec"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s := domain.NewSession("", "", req.TargetSec)

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

func getSessionStatus(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		s, ok := m.GetSession(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		resp := struct {
			ID        string `json:"id"`
			Completed bool   `json:"completed"`
			Current   int    `json:"currentStep"`
		}{
			ID:        s.ID,
			Completed: s.Completed,
			Current:   s.CurrentIdx,
		}

		json.NewEncoder(w).Encode(resp)
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

func startStep(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		idxStr := chi.URLParam(r, "idx")

		stepIdx, err := strconv.Atoi(idxStr)
		if err != nil {
			http.Error(w, "invalid step index", http.StatusBadRequest)
			return
		}

		if err := m.StartStep(id, stepIdx); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func stopStep(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		idxStr := chi.URLParam(r, "idx")

		stepIdx, err := strconv.Atoi(idxStr)
		if err != nil {
			http.Error(w, "invalid step index", http.StatusBadRequest)
			return
		}

		if err := m.StopStep(id, stepIdx); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
