package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sistemaCognitivo/internal/ai"
	"sistemaCognitivo/internal/memory"
	"strings"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	qrterminal "github.com/mdp/qrterminal/v3"
	"google.golang.org/protobuf/proto"

	_ "modernc.org/sqlite"
)

var (
	taskStore   *memory.Store
	client      *whatsmeow.Client
	mu          sync.RWMutex
	connected   bool
	mode        = "simulado"
	groupID     types.JID // ID do grupo "just do it"
	groupLoaded bool
)

// Init inicializa o módulo WhatsApp
func Init(store *memory.Store) error {
	taskStore = store

	ctx := context.Background()
	if err := os.MkdirAll("whatsapp_data", 0o755); err != nil {
		return fmt.Errorf("erro ao preparar pasta de sessao: %w", err)
	}

	dbLog := waLog.Stdout("WA-DB", "WARN", true)
	dsn := "file:whatsapp_data/store.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)"
	container, err := sqlstore.New(ctx, "sqlite", dsn, dbLog)
	if err != nil {
		return fmt.Errorf("erro ao abrir storage do whatsapp: %w", err)
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("erro ao obter device do whatsapp: %w", err)
	}

	clientLog := waLog.Stdout("WA", "INFO", true)
	client = whatsmeow.NewClient(device, clientLog)
	client.InitialAutoReconnect = true // reconecta em background se primeira tentativa falhar
	client.AddEventHandler(handleEvent)

	if client.Store.ID == nil {
		qrChan, err := client.GetQRChannel(ctx)
		if err != nil {
			return fmt.Errorf("erro ao obter canal de QR: %w", err)
		}

		go func() {
			for evt := range qrChan {
				switch evt.Event {
				case "code":
					fmt.Println("Escaneie o QR Code abaixo no WhatsApp (Dispositivos conectados):")
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)

					pngPath := filepath.Join("whatsapp_data", "qr.png")
					if err := qrcode.WriteFile(evt.Code, qrcode.Medium, 512, pngPath); err != nil {
						fmt.Printf("Falha ao gerar PNG do QR: %v\n", err)
					} else {
						fmt.Printf("QR salvo em: %s\n", pngPath)
						fmt.Println("Abra essa imagem e escaneie no WhatsApp > Dispositivos conectados")
					}
				default:
					fmt.Printf("Evento de pareamento WhatsApp: %s\n", evt.Event)
				}
			}
		}()
	}

	go func() {
		if err := client.Connect(); err != nil {
			fmt.Printf("⚠️ Falha na conexão inicial do WhatsApp, reconexão automática continuará: %v\n", err)
		}
	}()

	mu.Lock()
	mode = "real"
	mu.Unlock()

	fmt.Println("Módulo WhatsApp conectado")

	// Buscar grupo "just do it" em background
	go findGroupByName(ctx, "just do it")

	return nil
}

// IngestMessage simula a captura de uma mensagem e transforma em tarefa quando detectar compromisso.
func IngestMessage(message string) (memory.Task, error) {
	if taskStore == nil {
		return memory.Task{}, errors.New("whatsapp module not initialized")
	}

	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return memory.Task{}, errors.New("message is required")
	}

	task, err := ai.ExtractTaskFromText(trimmed)
	if err != nil {
		if errors.Is(err, ai.ErrNeedsContext) {
			return memory.Task{}, errors.New("compromisso vago; reformule com assunto e local. Ex.: Reunião com equipe de vendas no escritório hoje 15h")
		}
		if errors.Is(err, ai.ErrNotCommitment) {
			return memory.Task{}, errors.New("mensagem não parece compromisso; inclua data/horário e contexto de compromisso")
		}
		return memory.Task{}, err
	}

	task.Source = "whatsapp"
	stored := taskStore.Add(task)

	// Enviar confirmação para grupo "just do it" em background
	go func() {
		notification := buildCreatedTaskMessage(stored)
		if err := SendToGroup(context.Background(), notification); err != nil {
			fmt.Printf("⚠️ Falha ao notificar grupo: %v\n", err)
		} else {
			fmt.Printf("✅ Notificação enviada ao grupo para tarefa #%d\n", stored.ID)
		}
	}()

	return stored, nil
}

func handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		mu.Lock()
		connected = true
		mode = "real"
		mu.Unlock()
		fmt.Println("✅ WhatsApp conectado")
	case *events.Disconnected:
		mu.Lock()
		connected = false
		mu.Unlock()
		// WhatsMeow gerencia auto-reconnect internamente
		fmt.Println("⚠️ WhatsApp desconectado - aguardando reconexão automática...")
	case *events.Message:
		handleIncomingMessage(v)
	}
}

func handleIncomingMessage(evt *events.Message) {
	if taskStore == nil {
		return
	}

	var msg string
	if evt.Message.GetConversation() != "" {
		msg = evt.Message.GetConversation()
	} else if evt.Message.GetExtendedTextMessage() != nil {
		msg = evt.Message.GetExtendedTextMessage().GetText()
	}

	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}

	task, err := ai.ExtractTaskFromText(msg)
	if err != nil {
		if errors.Is(err, ai.ErrNeedsContext) {
			fmt.Printf("ℹ️ Mensagem rejeitada por falta de contexto: %q\n", msg)
			return
		}
		if errors.Is(err, ai.ErrNotCommitment) {
			fmt.Printf("ℹ️ Mensagem ignorada (não compromisso): %q\n", msg)
		}
		return
	}

	task.Source = "whatsapp"
	stored := taskStore.Add(task)

	// Enviar confirmação para grupo "just do it" em background
	go func() {
		notification := buildCreatedTaskMessage(stored)
		if err := SendToGroup(context.Background(), notification); err != nil {
			fmt.Printf("⚠️ Falha ao notificar grupo: %v\n", err)
		} else {
			fmt.Printf("✅ Grupo notificado sobre nova tarefa #%d\n", stored.ID)
		}
	}()
}

// GetStatus retorna status de conexão do módulo para o endpoint /status.
func GetStatus() map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()

	return map[string]interface{}{
		"connected": connected,
		"client":    client != nil,
		"mode":      mode,
		"group_found": groupLoaded,
	}
}

// findGroupByName busca grupo pelo nome e armazena ID
func findGroupByName(ctx context.Context, name string) {
	if client == nil {
		return
	}

	// Aguardar conexão (até 60s) antes de listar grupos
	for i := 0; i < 12; i++ {
		if client.IsConnected() {
			break
		}
		fmt.Printf("⏳ Aguardando WhatsApp conectar para buscar grupo (%d/12)...\n", i+1)
		time.Sleep(5 * time.Second)
	}

	if !client.IsConnected() {
		fmt.Println("⚠️ WhatsApp não conectou a tempo. Grupo será buscado ao reconectar.")
		return
	}

	groups, err := client.GetJoinedGroups(ctx)
	if err != nil {
		fmt.Printf("Erro ao listar grupos: %v\n", err)
		return
	}

	// Remove caracteres especiais/emojis e normaliza para comparação
	normalizeGroupName := func(s string) string {
		s = strings.ToLower(s)
		// Remove emojis e caracteres especiais
		result := ""
		for _, r := range s {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
				result += string(r)
			}
		}
		return strings.TrimSpace(result)
	}

	searchName := normalizeGroupName(name)

	for _, g := range groups {
		info, err := client.GetGroupInfo(ctx, g.JID)
		if err != nil {
			continue
		}

		if normalizeGroupName(info.Name) == searchName {
			mu.Lock()
			groupID = g.JID
			groupLoaded = true
			mu.Unlock()

			fmt.Printf("✅ Grupo '%s' encontrado e vinculado\n", info.Name)
			return
		}
	}

	fmt.Printf("⚠️ Grupo '%s' não encontrado. Envie uma mensagem no grupo para sincronizar.\n", name)
}

// SendToGroup envia mensagem para o grupo "just do it" com retry automático
func SendToGroup(ctx context.Context, message string) error {
	fmt.Printf("[WHATSAPP] SendToGroup called with message (first 50 chars): %.50s\n", message)
	
	mu.RLock()
	if !groupLoaded || groupID.IsEmpty() {
		mu.RUnlock()
		fmt.Printf("[WHATSAPP] Group not loaded, attempting to find...\n")
		findGroupByName(ctx, "just do it")
		mu.RLock()
		if !groupLoaded || groupID.IsEmpty() {
			mu.RUnlock()
			fmt.Printf("[WHATSAPP] ❌ Group not found after search\n")
			return errors.New("grupo não encontrado")
		}
	}
	gid := groupID
	fmt.Printf("[WHATSAPP] Group ID loaded: %s\n", gid.String())
	mu.RUnlock()

	if client == nil {
		fmt.Printf("[WHATSAPP] ❌ Client is nil\n")
		return errors.New("cliente WhatsApp não inicializado")
	}

	msg := &waE2E.Message{
		Conversation: proto.String(message),
	}

	// Retry até 5x aguardando reconexão automática do WhatsMeow
	delays := []time.Duration{3 * time.Second, 5 * time.Second, 10 * time.Second, 15 * time.Second, 20 * time.Second}
	var lastErr error
	for i, delay := range delays {
		connected := client.IsConnected()
		fmt.Printf("[WHATSAPP] Attempt %d/5: connected=%v\n", i+1, connected)
		
		if !connected {
			fmt.Printf("[WHATSAPP] ⏳ Waiting for reconnection... (delay: %v)\n", delay)
			time.Sleep(delay)
			continue
		}

		fmt.Printf("[WHATSAPP] Sending message to group %s...\n", gid.String())
		_, lastErr = client.SendMessage(ctx, gid, msg)
		if lastErr == nil {
			fmt.Printf("[WHATSAPP] ✅ Message sent successfully\n")
			return nil
		}

		fmt.Printf("[WHATSAPP] ⚠️ Attempt %d/5 failed: %v\n", i+1, lastErr)
		time.Sleep(delay)
	}

	fmt.Printf("[WHATSAPP] ❌ All 5 attempts failed\n")
	return fmt.Errorf("falha após 5 tentativas: %w", lastErr)
}

// SendReminderToGroup envia lembrete para o grupo
func SendReminderToGroup(task memory.Task) error {
	fmt.Printf("[WHATSAPP] SendReminderToGroup called for task #%d: %q\n", task.ID, task.Title)
	ctx := context.Background()
	msg := buildReminderMessage(task)
	fmt.Printf("[WHATSAPP] Sending reminder message: %s\n", msg)
	err := SendToGroup(ctx, msg)
	if err != nil {
		fmt.Printf("[WHATSAPP] ❌ SendReminderToGroup failed: %v\n", err)
	} else {
		fmt.Printf("[WHATSAPP] ✅ SendReminderToGroup succeeded\n")
	}
	return err
}

func buildCreatedTaskMessage(task memory.Task) string {
	msg := fmt.Sprintf("✅ Tarefa Criada!\n📌 %s\n⏰ %s", task.Title, task.DueAt.Format("02/01 15:04"))
	if needsMoreContext(task.Title) {
		msg += "\n❗ Falta contexto: assunto e local"
	}
	return msg
}

func buildReminderMessage(task memory.Task) string {
	msg := fmt.Sprintf("🔔 LEMBRETE\n📌 %s\n⏰ Prazo: %s", task.Title, task.DueAt.Format("02/01 15:04"))
	if needsMoreContext(task.Title) {
		msg += "\n❗ Detalhe melhor: de que e onde?"
	}
	return msg
}

func needsMoreContext(title string) bool {
	lower := strings.ToLower(strings.TrimSpace(title))
	if lower == "" {
		return true
	}

	padded := " " + lower + " "
	hasGenericTopic := hasAnyContains(lower, []string{"reuniao", "reunião", "consulta", "compromisso", "call", "meet"})
	hasDetailMarker := hasAnyContains(padded, []string{" com ", " sobre ", " no ", " na ", " em ", " para ", " com o ", " com a "})
	wordCount := len(strings.Fields(lower))

	// Ex.: "Reunião hoje 15h" -> genérico e sem contexto.
	return hasGenericTopic && !hasDetailMarker && wordCount <= 4
}

func hasAnyContains(text string, pieces []string) bool {
	for _, piece := range pieces {
		if strings.Contains(text, piece) {
			return true
		}
	}
	return false
}

// ListGroups retorna lista de grupos disponíveis para debug
func ListGroups(ctx context.Context) (map[string]interface{}, error) {
	if client == nil {
		return nil, errors.New("cliente WhatsApp não inicializado")
	}

	groups, err := client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar grupos: %w", err)
	}

	result := make([]map[string]string, 0)
	for _, g := range groups {
		info, err := client.GetGroupInfo(ctx, g.JID)
		if err != nil {
			continue
		}
		result = append(result, map[string]string{
			"name": info.Name,
			"jid":  g.JID.String(),
		})
	}

	return map[string]interface{}{
		"total":  len(result),
		"groups": result,
	}, nil
}
