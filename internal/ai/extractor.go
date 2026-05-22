package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"sistemaCognitivo/internal/memory"
	"sistemaCognitivo/internal/semantic"
)

var ErrNotCommitment = semantic.ErrNotCommitment
var ErrNeedsContext = semantic.ErrNeedsContext

const defaultSemanticServiceURL = "http://127.0.0.1:8090"

type SemanticAnalysis struct {
	Accepted     bool         `json:"accepted"`
	Task         *memory.Task `json:"task,omitempty"`
	Error        string       `json:"error,omitempty"`
	NeedsContext bool         `json:"needs_context,omitempty"`
}

// ExtractTaskFromText tenta usar o serviço semântico dedicado e cai para a análise local se necessário.
func ExtractTaskFromText(message string) (memory.Task, error) {
	if task, err := extractViaSemanticService(message); err == nil {
		return task, nil
	}

	return semantic.ExtractTaskFromText(message)
}

// AnalyzeSemantic consulta apenas o serviço semântico dedicado, sem fallback local.
func AnalyzeSemantic(message string) (SemanticAnalysis, error) {
	serviceURL := strings.TrimSpace(os.Getenv("SEMANTIC_SERVICE_URL"))
	if serviceURL == "" {
		serviceURL = defaultSemanticServiceURL
	}

	requestBody, err := json.Marshal(semanticAnalyzeRequest{Message: message})
	if err != nil {
		return SemanticAnalysis{}, err
	}

	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Post(serviceURL+"/analyze", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return SemanticAnalysis{}, err
	}
	defer response.Body.Close()

	var payload semanticAnalyzeResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return SemanticAnalysis{}, err
	}

	return SemanticAnalysis{
		Accepted:     payload.Accepted,
		Task:         payload.Task,
		Error:        payload.Error,
		NeedsContext: payload.NeedsContext,
	}, nil
}

type semanticAnalyzeRequest struct {
	Message string `json:"message"`
}

type semanticAnalyzeResponse struct {
	Accepted     bool         `json:"accepted"`
	Task         *memory.Task `json:"task,omitempty"`
	Error        string       `json:"error,omitempty"`
	NeedsContext bool         `json:"needs_context,omitempty"`
}

func extractViaSemanticService(message string) (memory.Task, error) {
	serviceURL := strings.TrimSpace(os.Getenv("SEMANTIC_SERVICE_URL"))
	if serviceURL == "" {
		serviceURL = defaultSemanticServiceURL
	}

	requestBody, err := json.Marshal(semanticAnalyzeRequest{Message: message})
	if err != nil {
		return memory.Task{}, err
	}

	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Post(serviceURL+"/analyze", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return memory.Task{}, err
	}
	defer response.Body.Close()

	var payload semanticAnalyzeResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return memory.Task{}, err
	}

	if payload.Accepted && payload.Task != nil {
		return *payload.Task, nil
	}

	switch {
	case errors.Is(mapRemoteError(payload.Error), ErrNotCommitment):
		return memory.Task{}, ErrNotCommitment
	case errors.Is(mapRemoteError(payload.Error), ErrNeedsContext):
		return memory.Task{}, ErrNeedsContext
	case payload.Error != "":
		return memory.Task{}, fmt.Errorf("%s", payload.Error)
	default:
		return memory.Task{}, ErrNotCommitment
	}
}

func mapRemoteError(message string) error {
	switch message {
	case ErrNotCommitment.Error():
		return ErrNotCommitment
	case ErrNeedsContext.Error():
		return ErrNeedsContext
	default:
		return errors.New(message)
	}
}
