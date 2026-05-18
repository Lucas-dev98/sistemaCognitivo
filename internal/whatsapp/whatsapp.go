package whatsapp

import (
	"errors"
	"fmt"
	"sistemaCognitivo/internal/ai"
	"sistemaCognitivo/internal/memory"
	"strings"
)

var taskStore *memory.Store

// Init inicializa o módulo WhatsApp
func Init(store *memory.Store) error {
	taskStore = store
	fmt.Println("Módulo WhatsApp iniciado (MVP)")
	return nil
}

// IngestMessage simula a captura de uma mensagem e transforma em tarefa quando detectar compromisso.
func IngestMessage(message string) (memory.Task, error) {
	if taskStore == nil {
		return memory.Task{}, errors.New("whatsapp module not initialized")
	}

	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return memory.Task{}, errors.New("message is required")
	}

	task, err := ai.ExtractTaskFromText(trimmed)
	if err != nil {
		return memory.Task{}, err
	}

	task.Source = "whatsapp"
	stored := taskStore.Add(task)
	return stored, nil
}

// GetStatus retorna status de conexão do módulo para o endpoint /status.
func GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"connected": false,
		"client":    false,
		"mode":      "simulado",
	}
}
