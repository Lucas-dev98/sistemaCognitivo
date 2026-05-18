package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sistemaCognitivo/internal/memory"
	"sistemaCognitivo/internal/scheduler"
	"sistemaCognitivo/internal/whatsapp"
)

type ingestRequest struct {
	Message string `json:"message"`
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

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/ingest/whatsapp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ingestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		task, err := whatsapp.IngestMessage(req.Message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

	log.Println("API MVP ouvindo em :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
