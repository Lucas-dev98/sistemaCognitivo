package ai

import (
	"errors"
	"testing"
	"time"
)

func TestExtractTaskFromText_CommitmentFiltering(t *testing.T) {
	tests := []struct {
		name       string
		msg        string
		shouldPass bool
		expectedErr error
	}{
		{name: "valid commitment with weekday", msg: "Reuniao com cliente sexta 9h", shouldPass: true},
		{name: "valid reminder", msg: "Lembrete pagar boleto amanha 10h", shouldPass: true},
		{name: "valid future intent", msg: "Vou ao dentista sexta 9h", shouldPass: true},
		{name: "valid imperative with weekday and day", msg: "9Finalizar estoque até quinta feira dia 21)", shouldPass: true},
		{name: "past statement", msg: "Eu fiz particular com Dr Breno de Mello Vitor", shouldPass: false, expectedErr: ErrNotCommitment},
		{name: "daily chatter", msg: "acordei agora", shouldPass: false, expectedErr: ErrNotCommitment},
		{name: "chat with temporal but no intent", msg: "caraca amor dormiu em sexta 9h", shouldPass: false, expectedErr: ErrNotCommitment},
		{name: "test phrase should be ignored", msg: "Teste lembrete sexta 9h", shouldPass: false, expectedErr: ErrNotCommitment},
		{name: "long context without schedule", msg: "Gerência Geral de Inovação. Temos várias frentes diferentes aqui na GG. Eu sou da VSI (Validação de Soluções de Inovação).", shouldPass: false, expectedErr: ErrNotCommitment},
		{name: "generic meeting without context", msg: "Reuniao hoje 15h", shouldPass: false, expectedErr: ErrNeedsContext},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ExtractTaskFromText(tt.msg)
			if tt.shouldPass && err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
			if !tt.shouldPass && !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestExtractTaskFromText_NormalizesTitle(t *testing.T) {
	task, err := ExtractTaskFromText("9Finalizar estoque até quinta feira dia 21)")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if task.Title != "Finalizar estoque até quinta feira dia 21" {
		t.Fatalf("unexpected normalized title: %q", task.Title)
	}
}

func TestExtractTaskFromText_PreservesMinutes(t *testing.T) {
	task, err := ExtractTaskFromText("Reuniao com cliente hoje 17:20")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if task.DueAt.Minute() != 20 {
		t.Fatalf("expected minute 20, got %d (due_at=%s)", task.DueAt.Minute(), task.DueAt.Format(time.RFC3339))
	}
}
