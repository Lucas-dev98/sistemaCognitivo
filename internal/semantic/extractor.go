package semantic

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sistemaCognitivo/internal/memory"
)

var (
	hourRegex    = regexp.MustCompile(`(?i)\b(\d{1,2})\s*h\b`)
	colonRegex   = regexp.MustCompile(`\b(\d{1,2}):(\d{2})\b`)
	dateDMRegex  = regexp.MustCompile(`\b(\d{1,2})/(\d{1,2})\b`)
	dayRegex     = regexp.MustCompile(`(?i)\bdia\s+(\d{1,2})\b`)
	spaceRegex   = regexp.MustCompile(`\s+`)
	leadingJunk  = regexp.MustCompile(`^[^\p{L}]+`)
	trailingJunk = regexp.MustCompile(`[\s\)\]\}\.,;:!\?]+$`)
)

var ErrNotCommitment = errors.New("mensagem não parece compromisso")
var ErrNeedsContext = errors.New("compromisso sem contexto suficiente")

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

var commitmentKeywords = []string{
	"reuniao", "reunião", "consulta", "medico", "médico", "dentista", "compromisso",
	"agenda", "agendar", "marcar", "marcado", "lembrete", "lembrar", "prazo",
	"entregar", "apresentar", "pagar", "vencimento", "call", "meet", "entrevista",
	"finalizar", "terminar", "concluir", "resolver", "enviar", "comprar", "renovar",
}

var futureIntentKeywords = []string{
	"amanha", "amanhã", "depois de amanha", "depois de amanhã", "hoje",
	"vou", "tenho", "preciso", "nao esquecer", "não esquecer", "lembra", "lembrar",
}

var noisePastKeywords = []string{
	"acordei", "acabei", "fiz", "fui", "comi", "almocei", "jant", "agora", "kkk",
}

var noiseChatKeywords = []string{
	"teste", "caraca", "amor", "dormiu", "oi", "bom dia", "boa tarde", "boa noite",
}

var noiseRefusalKeywords = []string{
	"infelizmente", "não vai dar", "nao vai dar", "oq eu consigo", "o que eu consigo",
	"consigo é", "o que posso", "posso é", "da pra", "dá pra", "não consigo", "nao consigo",
}

var actionIntentKeywords = []string{
	"tenho", "vou", "preciso", "marquei", "agendei", "agendar", "marcar",
	"lembrar", "lembrete", "nao esquecer", "não esquecer", "compromisso", "consulta", "reuniao", "reunião",
	"finalizar", "terminar", "concluir", "resolver", "enviar", "comprar", "renovar",
	"fazer", "preparar", "organizar",
}

// ExtractTaskFromText transforma texto livre em tarefa quando detectar compromisso.
func ExtractTaskFromText(message string) (memory.Task, error) {
	normalizedTitle := normalizeTaskTitle(message)
	lower := strings.ToLower(strings.TrimSpace(normalizedTitle))
	if lower == "" {
		return memory.Task{}, errors.New("empty message")
	}

	if !looksLikeCommitment(lower) {
		return memory.Task{}, ErrNotCommitment
	}

	hour := 9
	minute := 0
	hasExplicitTime := false
	if match := hourRegex.FindStringSubmatch(lower); len(match) == 2 {
		hasExplicitTime = true
		_, err := fmt.Sscanf(match[1], "%d", &hour)
		if err != nil || hour < 0 || hour > 23 {
			return memory.Task{}, errors.New("invalid hour")
		}
	} else if match := colonRegex.FindStringSubmatch(lower); len(match) == 3 {
		hasExplicitTime = true
		_, err := fmt.Sscanf(match[1], "%d", &hour)
		if err != nil || hour < 0 || hour > 23 {
			return memory.Task{}, errors.New("invalid hour")
		}
		_, err = fmt.Sscanf(match[2], "%d", &minute)
		if err != nil || minute < 0 || minute > 59 {
			return memory.Task{}, errors.New("invalid minute")
		}
	}

	now := time.Now()
	due := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	hasDateHint := false

	for token, wd := range weekdays {
		if strings.Contains(lower, token) {
			due = nextWeekday(now, wd, hour, minute)
			hasDateHint = true
			break
		}
	}

	if strings.Contains(lower, "amanha") || strings.Contains(lower, "amanhã") {
		t := now.Add(24 * time.Hour)
		due = time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, t.Location())
		hasDateHint = true
	}

	if strings.Contains(lower, "depois de amanha") || strings.Contains(lower, "depois de amanhã") {
		t := now.Add(48 * time.Hour)
		due = time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, t.Location())
		hasDateHint = true
	}

	if match := dateDMRegex.FindStringSubmatch(lower); len(match) == 3 {
		var day, month int
		if _, err := fmt.Sscanf(match[1], "%d", &day); err == nil {
			if _, err := fmt.Sscanf(match[2], "%d", &month); err == nil {
				candidate := time.Date(now.Year(), time.Month(month), day, hour, minute, 0, 0, now.Location())
				if candidate.Before(now) {
					candidate = candidate.AddDate(1, 0, 0)
				}
				due = candidate
				hasDateHint = true
			}
		}
	}

	if match := dayRegex.FindStringSubmatch(lower); len(match) == 2 {
		var day int
		if _, err := fmt.Sscanf(match[1], "%d", &day); err == nil && day >= 1 && day <= 31 {
			candidate := time.Date(now.Year(), now.Month(), day, hour, minute, 0, 0, now.Location())
			if candidate.Before(now) {
				candidate = candidate.AddDate(0, 1, 0)
			}
			due = candidate
			hasDateHint = true
		}
	}

	if strings.Contains(lower, "hoje") {
		if hasExplicitTime {
			due = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		} else {
			// Para "hoje" sem horário, agenda ainda hoje em uma janela próxima.
			due = defaultDueForToday(now)
		}
		hasDateHint = true
	}

	if !hasDateHint {
		return memory.Task{}, ErrNotCommitment
	}

	if due.Before(now) {
		due = due.Add(24 * time.Hour)
	}

	if lacksContext(normalizedTitle) {
		return memory.Task{}, ErrNeedsContext
	}

	return memory.Task{Title: normalizedTitle, DueAt: due}, nil
}

func lacksContext(title string) bool {
	lower := strings.ToLower(strings.TrimSpace(title))
	if lower == "" {
		return true
	}

	hasGenericTopic := hasAny(lower, []string{"reuniao", "reunião", "consulta", "compromisso", "call", "meet"})
	hasDetailMarker := hasAny(lower, []string{" com ", " sobre ", " no ", " na ", " em ", " para ", " com o ", " com a "})
	wordCount := len(strings.Fields(lower))

	return hasGenericTopic && !hasDetailMarker && wordCount <= 4
}

func normalizeTaskTitle(message string) string {
	normalized := strings.TrimSpace(message)
	normalized = spaceRegex.ReplaceAllString(normalized, " ")
	normalized = leadingJunk.ReplaceAllString(normalized, "")
	normalized = trailingJunk.ReplaceAllString(normalized, "")
	normalized = strings.TrimSpace(normalized)
	return normalized
}

func looksLikeCommitment(lower string) bool {
	score := 0

	hasTemporal := hasAny(lower, []string{
		"amanha", "amanhã", "depois de amanha", "depois de amanhã", "hoje",
		"segunda", "terça", "terca", "quarta", "quinta", "sexta", "sabado", "sábado", "domingo",
	}) || hourRegex.MatchString(lower) || colonRegex.MatchString(lower) || dateDMRegex.MatchString(lower) || dayRegex.MatchString(lower)

	if hasTemporal {
		score += 3
	}

	if hasAny(lower, commitmentKeywords) {
		score += 3
	}

	if hasAny(lower, futureIntentKeywords) {
		score += 2
	}

	hasActionIntent := hasAny(lower, actionIntentKeywords)
	if hasActionIntent {
		score += 2
	}

	if hasAny(lower, noisePastKeywords) && !hasTemporal {
		score -= 4
	}

	if hasAny(lower, noiseChatKeywords) {
		score -= 4
	}

	if hasAny(lower, noiseRefusalKeywords) {
		score -= 5
	}

	if hasAny(lower, []string{"levar", "entregar", "buscar", "encontrar"}) && hasAny(lower, noiseRefusalKeywords) {
		score -= 3
	}

	if len(strings.Fields(lower)) < 3 && !hasTemporal {
		score -= 2
	}

	if len(lower) > 120 && !hasAny(lower, commitmentKeywords) {
		score -= 3
	}

	hasStrongCommitment := hasAny(lower, commitmentKeywords)
	return hasTemporal && (hasActionIntent || hasStrongCommitment) && score >= 6
}

func hasAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func nextWeekday(now time.Time, wd time.Weekday, hour, minute int) time.Time {
	days := (int(wd) - int(now.Weekday()) + 7) % 7
	if days == 0 {
		days = 7
	}
	t := now.AddDate(0, 0, days)
	return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, t.Location())
}

func defaultDueForToday(now time.Time) time.Time {
	candidate := now.Add(30 * time.Minute)
	if candidate.Year() != now.Year() || candidate.Month() != now.Month() || candidate.Day() != now.Day() {
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())
	}

	return time.Date(now.Year(), now.Month(), now.Day(), candidate.Hour(), candidate.Minute(), 0, 0, now.Location())
}
