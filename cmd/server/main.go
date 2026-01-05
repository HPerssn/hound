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
	"github.com/hperssn/hound/internal/storage"
)

func main() {

	repo, err := storage.NewSQLiteRepository("./hound.db")
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer repo.Close()
	manager := runner.NewSessionManager()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/sessions", startSession(manager))
	r.Get("/sessions/{id}", getSession(manager))
	r.Post("/sessions/{id}/complete", completeSession(manager, repo))
	r.Post("/sessions/{id}/steps/{idx}/start", startStep(manager))
	r.Post("/sessions/{id}/steps/{idx}/stop", stopStep(manager))
	r.Post("/sessions/{id}/stop", stopSession(manager))
	r.Get("/sessions/{id}/events", httpapi.StreamSessionEvents(manager))
	r.Get("/sessions/{id}/status", getSessionStatus(manager))

	r.Get("/history", getHistory(repo))
	r.Get("/stats", getStats(repo))

	r.Get("/", serveIndex)
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func startSession(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TargetSec int `json:"targetSec"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.TargetSec <= 0 {
			respondError(w, "targetec must be positive", http.StatusBadRequest)
			return
		}

		session := domain.NewSession("", "", req.TargetSec)

		if err := m.StartSession(session); err != nil {
			respondError(w, err.Error(), http.StatusConflict)
			return
		}
		respondJSON(w, session, http.StatusCreated)
	}
}

func getSession(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		session, ok := m.GetSession(id)
		if !ok {
			respondError(w, "sessionnot found", http.StatusNotFound)
			return
		}

		respondJSON(w, session, http.StatusOK)
	}
}

func completeSession(m *runner.SessionManager, repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		var req struct {
			Success storage.SuccessLevel `json:"success"`
			Comment string               `json:"comment"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, ok := m.GetSession(id)
		if !ok {
			respondError(w, "session not found", http.StatusNotFound)
			return
		}

		record := storage.FromDomainSession(session, req.Success, req.Comment)
		if err := repo.SaveSession(record); err != nil {
			log.Printf("failed to save session: %v", err)
			respondError(w, "failed to save session", http.StatusInternalServerError)
			return
		}

		respondJSON(w, map[string]string{"status": "saved"}, http.StatusOK)
	}
}

func getSessionStatus(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		session, ok := m.GetSession(id)
		if !ok {
			respondError(w, "session not found", http.StatusNotFound)
			return
		}

		status := struct {
			ID        string `json:"id"`
			Completed bool   `json:"completed"`
			Current   int    `json:"currentStep"`
		}{
			ID:        session.ID,
			Completed: session.Completed,
			Current:   session.CurrentIdx,
		}

		respondJSON(w, status, http.StatusOK)
	}
}

func stopSession(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if err := m.StopSession(id); err != nil {
			respondError(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func startStep(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		stepIdx, err := parseStepIndex(r)

		if err != nil {
			respondError(w, "invalid step index", http.StatusBadRequest)
			return
		}

		if err := m.StartStep(id, stepIdx); err != nil {
			respondError(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func stopStep(m *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		stepIdx, err := parseStepIndex(r)
		if err != nil {
			respondError(w, "invalid step index", http.StatusBadRequest)
			return
		}

		if err := m.StopStep(id, stepIdx); err != nil {
			respondError(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getHistory(repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		if userID == "" {
			userID = "" // Default to empty user for now
		}

		sessions, err := repo.GetSessionsByUser(userID)
		if err != nil {
			log.Printf("Failed to get history: %v", err)
			respondError(w, "failed to retrieve history", http.StatusInternalServerError)
			return
		}

		respondJSON(w, sessions, http.StatusOK)
	}
}

func getStats(repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		if userID == "" {
			userID = "" // Default to empty user for now
		}

		stats, err := repo.GetSessionStats(userID)
		if err != nil {
			log.Printf("Failed to get stats: %v", err)
			respondError(w, "failed to retrieve stats", http.StatusInternalServerError)
			return
		}

		respondJSON(w, stats, http.StatusOK)
	}
}

// helpers below
func parseStepIndex(r *http.Request) (int, error) {
	idxStr := chi.URLParam(r, "idx")
	return strconv.Atoi(idxStr)
}

func respondJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
