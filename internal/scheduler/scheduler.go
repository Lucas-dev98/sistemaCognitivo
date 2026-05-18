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
			for _, task := range tasks {
				if task.Reminded {
					continue
				}

				delta := task.DueAt.Sub(now)
				if delta <= 30*time.Minute && delta >= 0 {
					reminders.Notify(task)
					store.MarkReminded(task.ID)
					log.Printf("lembrete registrado para tarefa #%d", task.ID)
				}
			}
		}
	}()
}
