package memory

import (
	"sync"
	"time"
)

// Task representa um compromisso detectado pela camada de captura + IA.
type Task struct {
	ID       int       `json:"id"`
	Source   string    `json:"source"`
	Title    string    `json:"title"`
	DueAt    time.Time `json:"due_at"`
	Created  time.Time `json:"created_at"`
	Reminded bool      `json:"reminded"`
}

// TaskStore define operações de persistência de tarefas.
type TaskStore interface {
	Add(task Task) Task
	List() []Task
	MarkReminded(id int)
}

// StoreHealth representa o estado do backend de persistencia de tarefas.
type StoreHealth struct {
	Backend   string `json:"backend"`
	Persisted bool   `json:"persisted"`
	Healthy   bool   `json:"healthy"`
	Error     string `json:"error,omitempty"`
}

// HealthAwareStore expõe estado operacional do backend de tarefas.
type HealthAwareStore interface {
	HealthStatus() StoreHealth
}

// Store e um armazenamento em memoria para o MVP.
type Store struct {
	mu     sync.RWMutex
	nextID int
	tasks  []Task
}

func NewStore() *Store {
	return &Store{nextID: 1, tasks: make([]Task, 0)}
}

func (s *Store) Add(task Task) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.ID = s.nextID
	s.nextID++
	task.Created = time.Now()
	s.tasks = append(s.tasks, task)
	return task
}

func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Task, len(s.tasks))
	copy(out, s.tasks)
	return out
}

func (s *Store) MarkReminded(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Reminded = true
			return
		}
	}
}

func (s *Store) HealthStatus() StoreHealth {
	return StoreHealth{
		Backend:   "memory",
		Persisted: false,
		Healthy:   true,
	}
}
