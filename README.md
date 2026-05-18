# Assistente Cognitivo Pessoal com IA

## Visão Geral

Sistema para reduzir esquecimentos, centralizar compromissos, transformar mensagens/documentos em lembretes e atuar como memória externa inteligente.

### Módulos
- Capture Layer (WhatsApp, Email, Arquivos)
- Queue Layer (Redis)
- AI Engine (Ollama, Qwen 2.5 7B, Qwen2.5-VL)
- OCR Engine (Tesseract)
- Memory Engine (PostgreSQL, pgvector)
- Reminder Engine (gocron)
- Notification System (WhatsApp)

### Estrutura de Pastas
- cmd/
  - main.go (servidor HTTP + orquestrador)
  - worker-ai/ (processador IA paralelo)
  - worker-vision/ (processador de imagens)
- internal/
  - whatsapp/ (integração WhatsMeow com autenticação QR)
  - email/
  - ai/
  - ocr/
  - vision/
  - reminders/
  - scheduler/
  - memory/
  - queue/
  - workers/
  - database/

### Infraestrutura
- Docker Compose: bot-go, redis, postgres, ollama, ocr, worker-ai, worker-vision

## 🚀 MVP Atual - WhatsApp + Lembretes

O MVP implementado inclui:
- ✅ Integração real com WhatsApp (WhatsMeow + autenticação QR)
- ✅ Ingestão de mensagem em tempo real
- ✅ Extração simples de compromisso (dia da semana/amanhã + horário)
- ✅ Armazenamento em memória com estado
- ✅ Scheduler com verificação periódica
- ✅ Lembretes com confirmação automática
- ✅ API HTTP para testes

### Como Iniciar - Opção 1: WhatsApp Real

```bash
# 1. Instalar dependências
go mod tidy

# 2. Rodar aplicação
go run ./cmd/main.go

# 3. Escanear QR Code com seu WhatsApp
#    (Configurações → Dispositivos conectados → Conectar dispositivo)

# 4. Envie uma mensagem no WhatsApp
"Sexta tenho entrevista às 14h"

# 5. Você receberá confirmação automática!
```

Para mais detalhes: [WHATSAPP_SETUP.md](WHATSAPP_SETUP.md)

### Como Iniciar - Opção 2: API HTTP (Teste Local)

```bash
# 1. Rodar aplicação
go run ./cmd/main.go

# 2. Em outro terminal, envie mensagem simulada
curl -X POST http://localhost:8080/ingest/whatsapp \
  -H "Content-Type: application/json" \
  -d '{"message":"sexta tenho prova as 19h"}'

# 3. Listar tarefas criadas
curl http://localhost:8080/tasks

# 4. Ver status da conexão
curl http://localhost:8080/status
```

### Endpoints Disponíveis

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/health` | GET | Verifica saúde da API |
| `/ingest/whatsapp` | POST | Ingesta manual de mensagem |
| `/tasks` | GET | Lista todas as tarefas |
| `/status` | GET | Status da conexão WhatsApp |

## 📱 Integração WhatsApp (WhatsMeow)

### Como Funciona

1. **Autenticação**: Escaneia QR Code com seu celular
2. **Monitoramento**: App fica online recebendo mensagens em tempo real
3. **Análise**: Cada mensagem é automaticamente processada para detectar compromissos
4. **Persistência**: Sessão fica armazenada (não é preciso fazer login novamente)

### Fluxo Real

```
Você no WhatsApp:
"Sexta tenho entrevista às 14h"
        ↓
Servidor detecta mensagem
        ↓
IA extrai: dia=sexta, hora=14h
        ↓
Task criada e armazenada
        ↓
Resposta automática:
"✅ Tarefa criada!
 📌 Sexta tenho entrevista às 14h
 ⏰ 22/05 14:00"
```

## 🗓️ Roadmap

### MVP - Fase 1 ✅
- [x] Integração WhatsApp (WhatsMeow)
- [x] Extração básica de compromissos
- [x] Armazenamento em memória
- [x] Scheduler de lembretes
- [x] API HTTP

### Fase 2 - Email + OCR
- [ ] Integração Gmail
- [ ] Processamento de PDFs
- [ ] OCR com Tesseract
- [ ] Extração de imagens (vision AI)

### Fase 3 - Memória Inteligente
- [ ] PostgreSQL + pgvector
- [ ] Busca semântica
- [ ] Contexto persistente
- [ ] Relacionamento entre tarefas

### Fase 4 - Dashboard
- [ ] Interface React
- [ ] Analytics em tempo real
- [ ] Modo foco
- [ ] Integração calendário

## 🛠️ Ferramentas e Tecnologias

| Componente | Tecnologia | Status |
|-----------|-----------|--------|
| Linguagem | Go 1.22 | ✅ |
| WhatsApp | WhatsMeow | ✅ |
| IA (NLP) | Qwen 2.5 7B (Ollama) | 🔜 |
| IA (Vision) | Qwen2.5-VL | 🔜 |
| OCR | Tesseract | 🔜 |
| Queue | Redis | 🔜 |
| Banco | PostgreSQL + pgvector | 🔜 |
| Scheduler | gocron | ✅ |
| API | HTTP standard lib | ✅ |
| Container | Docker Compose | 🔜 |

## 📚 Documentação

- [WHATSAPP_SETUP.md](WHATSAPP_SETUP.md) - Guia completo de integração WhatsApp
- [docker-compose.yml](docker-compose.yml) - Configuração de containers
- [go.mod](go.mod) - Dependências do projeto

## 🚨 Troubleshooting

**P: Diz que WhatsApp não conecta**
R: Seu número pode estar banido. Tente com outro WhatsApp ou use modo API.

**P: Como pausar/desconectar?**
R: Pressione Ctrl+C no terminal. Próxima execução reconecta automaticamente.

**P: Posso usar em múltiplos números?**
R: Faça delete de `whatsapp_data/store.db` para fazer logout e escanear novo QR.

## 📞 Próximos Passos

1. Integrar **Redis** para fila assíncrona
2. Conectar **Ollama** para IA local avançada
3. Implementar **PostgreSQL** para persistência real
4. Criar **Dashboard React** para visualização
5. Adicionar **Gmail** para emails automáticos
