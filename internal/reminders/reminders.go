package reminders

import (
	"fmt"
	"log"
	"sistemaCognitivo/internal/memory"
)

// NotifyFunc é um callback para função de notificação customizada (ex: WhatsApp)
var NotifyFunc func(task memory.Task) error

// Notify emite aviso no console e via callback se disponível.
func Notify(task memory.Task) {
	fmt.Printf("[LEMBRETE] %s | prazo: %s | origem: %s\n", task.Title, task.DueAt.Format("02/01 15:04"), task.Source)
	log.Printf("[REMINDERS] Notifying task #%d: %q", task.ID, task.Title)

	// Tentar enviar via callback (ex: grupo WhatsApp)
	if NotifyFunc != nil {
		log.Printf("[REMINDERS] Calling NotifyFunc callback for task #%d", task.ID)
		if err := NotifyFunc(task); err != nil {
			log.Printf("[REMINDERS] ❌ Erro ao enviar notificação customizada: %v", err)
		} else {
			fmt.Printf("[REMINDERS] ✅ Lembrete enviado ao grupo com sucesso\n")
			log.Printf("[REMINDERS] ✅ NotifyFunc succeeded for task #%d", task.ID)
		}
	} else {
		log.Printf("[REMINDERS] ⚠️ NotifyFunc not set, task #%d cannot be sent to group", task.ID)
	}
}
