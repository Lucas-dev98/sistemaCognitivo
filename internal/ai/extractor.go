package ai

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sistemaCognitivo/internal/memory"
)

var hourRegex = regexp.MustCompile(`(?i)(\d{1,2})\s*h`)

var weekdays = map[string]time.Weekday{
	"domingo": time.Sunday,
	"segunda": time.Monday,
	"terca":   time.Tuesday,
	"terça":   time.Tuesday,
	"quarta":  time.Wednesday,
	"quinta":  time.Thursday,
	"sexta":   time.Friday,
	"sabado":  time.Saturday,
	"sábado":  time.Saturday,
}

// ExtractTaskFromText faz extracao simples de compromisso em texto livre (MVP).
func ExtractTaskFromText(message string) (memory.Task, error) {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return memory.Task{}, errors.New("empty message")
	}

	hour := 9
	if match := hourRegex.FindStringSubmatch(lower); len(match) == 2 {
		_, err := fmt.Sscanf(match[1], "%d", &hour)
		if err != nil || hour < 0 || hour > 23 {
			return memory.Task{}, errors.New("invalid hour")
		}
	}

	now := time.Now()
	due := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())

	for token, wd := range weekdays {
		if strings.Contains(lower, token) {
			due = nextWeekday(now, wd, hour)
			break
		}
	}

	if strings.Contains(lower, "amanha") || strings.Contains(lower, "amanhã") {
		t := now.Add(24 * time.Hour)
		due = time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	}

	if due.Before(now) {
		due = due.Add(24 * time.Hour)
	}

	return memory.Task{Title: message, DueAt: due}, nil
}

func nextWeekday(now time.Time, wd time.Weekday, hour int) time.Time {
	days := (int(wd) - int(now.Weekday()) + 7) % 7
	if days == 0 {
		days = 7
	}
	t := now.AddDate(0, 0, days)
	return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
}
