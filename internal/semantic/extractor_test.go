package semantic

import (
	"testing"
	"time"
)

func TestExtractTaskFromText_TodayWithoutHourStaysToday(t *testing.T) {
	now := time.Now()
	task, err := ExtractTaskFromText("Fazer um bolo hoje")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	y1, m1, d1 := now.Date()
	y2, m2, d2 := task.DueAt.Date()
	if y1 != y2 || m1 != m2 || d1 != d2 {
		t.Fatalf("expected due date today (%02d/%02d/%d), got %02d/%02d/%d", d1, m1, y1, d2, m2, y2)
	}

	if task.DueAt.Before(now) {
		t.Fatalf("expected due time in the future for today, got %s (now=%s)", task.DueAt.Format(time.RFC3339), now.Format(time.RFC3339))
	}
}
