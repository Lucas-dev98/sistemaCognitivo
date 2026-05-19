package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sistemaCognitivo/internal/memory"
	"sistemaCognitivo/internal/reminders"
	"sistemaCognitivo/internal/scheduler"
	"sistemaCognitivo/internal/whatsapp"
)

type ingestRequest struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: message})
}

func main() {
	fmt.Println("Assistente Cognitivo Pessoal iniciado.")
	store := memory.NewStore()
	scheduler.Start(store)

	// Tentar conectar ao WhatsApp
	if err := whatsapp.Init(store); err != nil {
		log.Printf("⚠️ WhatsApp não conectado: %v\n", err)
		log.Println("API funcionará apenas em modo simulado (POST /ingest/whatsapp)")
	}

	// Configurar callback de notificação para grupo WhatsApp
	reminders.NotifyFunc = whatsapp.SendReminderToGroup

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/ingest/whatsapp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req ingestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		task, err := whatsapp.IngestMessage(req.Message)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(task)
	})

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(store.List())
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(whatsapp.GetStatus())
	})

	http.HandleFunc("/debug/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		groups, err := whatsapp.ListGroups(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(groups)
	})

	log.Println("API MVP ouvindo em :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
