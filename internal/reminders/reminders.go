package reminders

import (
	"fmt"
	"sistemaCognitivo/internal/memory"
)

// Notify emite aviso no console; no futuro sera enviado ao WhatsApp privado.
func Notify(task memory.Task) {
	fmt.Printf("[LEMBRETE] %s | prazo: %s | origem: %s\n", task.Title, task.DueAt.Format("02/01 15:04"), task.Source)
}
