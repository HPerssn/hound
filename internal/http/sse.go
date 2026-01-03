package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hperssn/hound/internal/runner"
)

func StreamSessionEvents(manager *runner.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		events, ok := manager.Events(id)
		if !ok {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		for {
			select {
			case step, ok := <-events:
				if !ok {
					return
				}

				data, _ := json.Marshal(step)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()

			case <-r.Context().Done():
				return
			}
		}
	}
}
