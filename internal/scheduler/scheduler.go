package scheduler

import (
	"log"
	"time"

	"sistemaCognitivo/internal/memory"
	"sistemaCognitivo/internal/reminders"
)

// Start inicia um loop simples que dispara lembretes de itens proximos do prazo.
func Start(store *memory.Store) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			tasks := store.List()
			checked := 0
			reminded := 0
			
			for _, task := range tasks {
				checked++
				if task.Reminded {
					continue
				}

				delta := task.DueAt.Sub(now)
				log.Printf("[SCHEDULER] Tarefa #%d: prazo em %.0f minutos (próximo: %v)", task.ID, delta.Minutes(), delta <= 30*time.Minute && delta >= 0)
				
				if delta <= 30*time.Minute && delta >= 0 {
					log.Printf("[SCHEDULER] ⏰ Disparando lembrete para tarefa #%d: %q (prazo: %s)", task.ID, task.Title, task.DueAt.Format("02/01 15:04"))
					reminders.Notify(task)
					store.MarkReminded(task.ID)
					reminded++
					log.Printf("[SCHEDULER] ✅ Lembrete registrado para tarefa #%d", task.ID)
				}
			}
			
			log.Printf("[SCHEDULER] Ciclo completo: %d tarefas verificadas, %d lembretes disparados", checked, reminded)
		}
	}()
}
