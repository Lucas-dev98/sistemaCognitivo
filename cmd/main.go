package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sistemaCognitivo/internal/ai"
	"sistemaCognitivo/internal/memory"
	"sistemaCognitivo/internal/reminders"
	"sistemaCognitivo/internal/scheduler"
	"sistemaCognitivo/internal/whatsapp"
	"strings"
)

type ingestRequest struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type semanticPreviewRequest struct {
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func resolvePostgresDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn
	}

	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	if host == "" {
		host = "localhost"
	}

	port := strings.TrimSpace(os.Getenv("DB_PORT"))
	if port == "" {
		port = "5432"
	}

	user := strings.TrimSpace(os.Getenv("DB_USER"))
	if user == "" {
		user = "user"
	}

	password := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	if password == "" {
		password = "password"
	}

	name := strings.TrimSpace(os.Getenv("DB_NAME"))
	if name == "" {
		name = "cognitive"
	}

	sslmode := strings.TrimSpace(os.Getenv("DB_SSLMODE"))
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslmode)
}

func initTaskStore() memory.TaskStore {
	dsn := resolvePostgresDSN()
	store, err := memory.NewPostgresStore(dsn)
	if err != nil {
		log.Printf("⚠️ PostgreSQL indisponível, usando memória: %v", err)
		return memory.NewStore()
	}

	log.Println("✅ Store de tarefas em PostgreSQL ativado")
	return store
}

func main() {
	fmt.Println("Assistente Cognitivo Pessoal iniciado.")
	store := initTaskStore()
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

	http.HandleFunc("/health/db", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		health, ok := store.(memory.HealthAwareStore)
		if !ok {
			writeJSON(w, http.StatusOK, memory.StoreHealth{
				Backend:   "unknown",
				Persisted: false,
				Healthy:   false,
				Error:     "store sem health check",
			})
			return
		}

		status := health.HealthStatus()
		httpCode := http.StatusOK
		if !status.Healthy {
			httpCode = http.StatusServiceUnavailable
		}

		writeJSON(w, httpCode, status)
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

		waStatus := whatsapp.GetStatus()
		health, ok := store.(memory.HealthAwareStore)
		if ok {
			dbStatus := health.HealthStatus()
			waStatus["store_backend"] = dbStatus.Backend
			waStatus["store_persisted"] = dbStatus.Persisted
			waStatus["store_healthy"] = dbStatus.Healthy
			if dbStatus.Error != "" {
				waStatus["store_error"] = dbStatus.Error
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(waStatus)
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

	http.HandleFunc("/semantic/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req semanticPreviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		analysis, err := ai.AnalyzeSemantic(req.Message)
		if err != nil {
			writeJSONError(w, http.StatusBadGateway, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, analysis)
	})

	if _, err := os.Stat("web"); err == nil {
		webFS := http.FileServer(http.Dir("web"))
		http.Handle("/styles.css", webFS)
		http.Handle("/app.js", webFS)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, "web/index.html")
		})
		log.Println("Frontend disponível em /")
	}

	log.Println("API MVP ouvindo em :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
