package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sistemaCognitivo/internal/memory"
	"sistemaCognitivo/internal/semantic"
)

type analyzeRequest struct {
	Message string `json:"message"`
}

type analyzeResponse struct {
	Accepted     bool         `json:"accepted"`
	Task         *memory.Task `json:"task,omitempty"`
	Error        string       `json:"error,omitempty"`
	NeedsContext bool         `json:"needs_context,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func main() {
	port := os.Getenv("SEMANTIC_SERVICE_PORT")
	if port == "" {
		port = "8090"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	http.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, analyzeResponse{Accepted: false, Error: "method not allowed"})
			return
		}

		var req analyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, analyzeResponse{Accepted: false, Error: "invalid json body"})
			return
		}

		task, err := semantic.ExtractTaskFromText(req.Message)
		if err != nil {
			status := http.StatusBadRequest
			if err == semantic.ErrNotCommitment {
				status = http.StatusUnprocessableEntity
			}
			if err == semantic.ErrNeedsContext {
				status = http.StatusUnprocessableEntity
			}

			writeJSON(w, status, analyzeResponse{
				Accepted:     false,
				Error:        err.Error(),
				NeedsContext: err == semantic.ErrNeedsContext,
			})
			return
		}

		writeJSON(w, http.StatusOK, analyzeResponse{
			Accepted: true,
			Task:     &task,
		})
	})

	addr := ":" + port
	log.Printf("Serviço semântico ouvindo em %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
