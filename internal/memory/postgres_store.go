package memory

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// PostgresStore persiste tarefas em PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexao postgres: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("erro ao conectar postgres: %w", err)
	}

	if err := ensureSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func ensureSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id BIGSERIAL PRIMARY KEY,
		source TEXT NOT NULL,
		title TEXT NOT NULL,
		due_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		reminded BOOLEAN NOT NULL DEFAULT FALSE
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_due_at ON tasks(due_at);
	CREATE INDEX IF NOT EXISTS idx_tasks_reminded_due_at ON tasks(reminded, due_at);
	`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("erro ao criar schema postgres: %w", err)
	}
	return nil
}

func (s *PostgresStore) Add(task Task) Task {
	created := time.Now()
	query := `
		INSERT INTO tasks (source, title, due_at, created_at, reminded)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	if err := s.db.QueryRow(query, task.Source, task.Title, task.DueAt, created, task.Reminded).Scan(&task.ID); err != nil {
		log.Printf("[STORE] falha ao inserir tarefa no postgres, mantendo tarefa sem ID persistido: %v", err)
		task.Created = created
		return task
	}

	task.Created = created
	return task
}

func (s *PostgresStore) List() []Task {
	rows, err := s.db.Query(`
		SELECT id, source, title, due_at, created_at, reminded
		FROM tasks
		ORDER BY id ASC
	`)
	if err != nil {
		log.Printf("[STORE] falha ao listar tarefas no postgres: %v", err)
		return []Task{}
	}
	defer rows.Close()

	out := make([]Task, 0)
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Source, &task.Title, &task.DueAt, &task.Created, &task.Reminded); err != nil {
			log.Printf("[STORE] falha ao ler tarefa do postgres: %v", err)
			continue
		}
		out = append(out, task)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[STORE] erro de cursor ao listar tarefas: %v", err)
	}

	return out
}

func (s *PostgresStore) MarkReminded(id int) {
	if _, err := s.db.Exec(`UPDATE tasks SET reminded = TRUE WHERE id = $1`, id); err != nil {
		log.Printf("[STORE] falha ao marcar lembrete no postgres para tarefa #%d: %v", id, err)
	}
}

func (s *PostgresStore) HealthStatus() StoreHealth {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return StoreHealth{
			Backend:   "postgres",
			Persisted: true,
			Healthy:   false,
			Error:     err.Error(),
		}
	}

	return StoreHealth{
		Backend:   "postgres",
		Persisted: true,
		Healthy:   true,
	}
}
