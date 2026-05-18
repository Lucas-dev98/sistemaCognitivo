# Fluxo Completo da Aplicação - Assistente Cognitivo Pessoal

## 🎯 Resumo Executivo

A aplicação funciona em **3 camadas principais**:

```
1️⃣ CAPTURA
   ↓ (Mensagem chega)
2️⃣ PROCESSAMENTO
   ↓ (IA analisa)
3️⃣ NOTIFICAÇÃO
   ↓ (Lembrete dispara)
```

---

## 📊 Diagrama Completo

```
┌─────────────────────────────────────────────────────────────────┐
│                    ASSISTENTE COGNITIVO                          │
└─────────────────────────────────────────────────────────────────┘

1. CAPTURA LAYER
   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
   │  WhatsApp    │  │    Email     │  │   PDFs       │
   │  (WhatsMeow) │  │  (Gmail API) │  │ (Tesseract)  │
   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                            ↓
2. QUEUE LAYER
   ┌────────────────────────────────┐
   │         Redis Queue            │
   │  (Processa de forma assíncrona)│
   └────────────┬───────────────────┘
                ↓
3. AI ENGINE
   ┌──────────────────────────────────┐
   │  • Qwen 2.5 7B (NLP)             │
   │  • Qwen2.5-VL (Vision)           │
   │  • Tesseract OCR                 │
   └────────────┬─────────────────────┘
                ↓
4. MEMORY ENGINE
   ┌──────────────────────────────────┐
   │  • PostgreSQL (dados)            │
   │  • pgvector (busca semântica)    │
   │  • Cache em memória (MVP)        │
   └────────────┬─────────────────────┘
                ↓
5. SCHEDULER ENGINE
   ┌──────────────────────────────────┐
   │  • gocron (agendamento)          │
   │  • Verifica prazos               │
   │  • Dispara lembretes             │
   └────────────┬─────────────────────┘
                ↓
6. NOTIFICATION SYSTEM
   ┌──────────────────────────────────┐
   │  WhatsApp Privado                │
   │  • Agenda diária                 │
   │  • Alertas urgentes              │
   │  • Resumos                       │
   └──────────────────────────────────┘
```

---

## 🔄 Fluxo Passo a Passo - Exemplo Real

### Cenário: Usuário recebe edital de concurso no WhatsApp

```
┌─────────────────────────────────────────────────────────────┐
│ INÍCIO: Usuário escreve no WhatsApp                         │
│                                                             │
│ "Edital de TI saiu! Inscrições até 30/06. Vou participar" │
└──────────────────────────┬──────────────────────────────────┘

                           ↓

┌──────────────────────────────────────────────────────────────┐
│ ETAPA 1: CAPTURA - WhatsMeow intercepta mensagem            │
│                                                              │
│ • Conecta ao WhatsApp via protocolo reverso                │
│ • Autentica via QR Code                                    │
│ • Recebe mensagem em tempo real                            │
│ • Verifica: não é mensagem de bot, não é da gente         │
└──────────────────────────┬─────────────────────────────────┘

                           ↓

┌──────────────────────────────────────────────────────────────┐
│ ETAPA 2: FILA - Redis enfileira para processamento          │
│                                                              │
│ Fila: [                                                     │
│   {msg: "Edital...", source: "whatsapp", ts: ...},         │
│   ...                                                       │
│ ]                                                           │
└──────────────────────────┬─────────────────────────────────┘

                           ↓

┌──────────────────────────────────────────────────────────────┐
│ ETAPA 3: AI ENGINE - Qwen 2.5 7B analisa                   │
│                                                              │
│ NLP detecciona:                                             │
│ • Tipo: Concurso/Inscrição/Edital                          │
│ • Prazo: 30 de junho                                       │
│ • Prioridade: ALTA (você mencionou que vai participar)    │
│ • Contexto: Você buscava oportunidade em TI               │
│                                                              │
│ Resultado: Task criada                                      │
│ {                                                           │
│   id: 1,                                                    │
│   title: "Edital de TI - inscrição até 30/06",            │
│   due_at: "2026-06-30T23:59:00",                          │
│   priority: "HIGH",                                        │
│   source: "whatsapp",                                      │
│   context: ["edital", "ti", "concurso"]                   │
│ }                                                           │
└──────────────────────────┬─────────────────────────────────┘

                           ↓

┌──────────────────────────────────────────────────────────────┐
│ ETAPA 4: MEMORY ENGINE - Armazena                          │
│                                                              │
│ Opção A (MVP - Atual):                                      │
│ └─ Memória RAM do servidor                                 │
│                                                              │
│ Opção B (Futuro):                                           │
│ ├─ PostgreSQL (dados estruturados)                         │
│ ├─ pgvector (embedding semântico)                         │
│ └─ Busca inteligente: "Quando vence edital?"              │
└──────────────────────────┬─────────────────────────────────┘

                           ↓

┌──────────────────────────────────────────────────────────────┐
│ ETAPA 5: CONFIRMAÇÃO - Envia resposta automática            │
│                                                              │
│ WhatsApp privado recebe:                                    │
│ "✅ Tarefa criada!                                          │
│  📌 Edital de TI - inscrição até 30/06                    │
│  ⏰ 30/06/2026                                             │
│  ⚠️ ALTA PRIORIDADE"                                       │
└──────────────────────────┬─────────────────────────────────┘

                           ↓

┌──────────────────────────────────────────────────────────────┐
│ ETAPA 6: SCHEDULER - Monitora                              │
│                                                              │
│ gocron.Every(1h).Do(CheckUpcomingTasks)                    │
│                                                              │
│ Checks:                                                     │
│ • 07/06: -23 dias → Notificação silenciosa                │
│ • 25/06: -5 dias → Alerta: "Faltam 5 dias"               │
│ • 29/06: -1 dia → URGENTE: "Última chance amanhã!"        │
│ • 30/06 08h: -15h → CRÍTICO: "Faltam 16 horas!"          │
│ • 30/06 23h: -1h → FINAL: "Última hora para inscrever!"  │
└──────────────────────────┬─────────────────────────────────┘

                           ↓

┌──────────────────────────────────────────────────────────────┐
│ ETAPA 7: NOTIFICAÇÃO - WhatsApp Privado                    │
│                                                              │
│ Você recebe progressivamente:                              │
│                                                              │
│ [07/06] 🔔 Edital aberto para inscrição (TI)             │
│ [25/06] ⏰ 5 dias restantes - Edital TI                  │
│ [29/06] 🚨 URGENTE - Falta 1 dia para inscrever          │
│ [30/06] 🔴 CRÍTICO - Últimas 16 horas! (8x por dia)      │
│                                                              │
│ Com informações úteis:                                      │
│ • Link para edital                                         │
│ • Documentos necessários                                   │
│ • Checklist de inscrição                                  │
└──────────────────────────┬─────────────────────────────────┘

                           ↓

┌──────────────────────────────────────────────────────────────┐
│ FIM: Você não esquece!                                      │
│                                                              │
│ ✅ Consegue se inscrever no edital                          │
│ ✅ Participa do concurso                                    │
│ ✅ Oportunidade não é perdida                               │
└──────────────────────────────────────────────────────────────┘
```

---

## 🧠 Tipos de Análise que o Sistema Faz

### Extração de Datas (MVP Atual)

```
Entrada: "Sexta tenho entrevista as 14h"
         ↓
Processamento:
  • Regex: \d{1,2}\s*h → Encontra "14"
  • Weekday map: "sexta" → Friday
  • Calcula próxima sexta
  ↓
Saída: due_at = 2026-05-22T14:00:00
```

### Análise Avançada (IA Local - Futuro)

```
Entrada: "Tenho reunião com João amanhã no Zoom as 3 da tarde"
         ↓
Processamento (Qwen 2.5 7B):
  • NLP: Detecta "reunião", "João", "Zoom", "3 da tarde"
  • Contexto: João = colega do projeto X
  • Relaciona com tarefas antigas
  ↓
Saída: 
{
  type: "meeting",
  participants: ["João"],
  platform: "Zoom",
  time: "15:00",
  related_tasks: [5, 12, 17]
}
```

### Análise de Contexto (Futuro com pgvector)

```
Pergunta: "Quando vence o edital?"
          ↓
pgvector busca:
  • Semelhança semântica com "edital"
  • Encontra: "Edital de TI - inscrição até 30/06"
  ↓
Resposta: "O edital vence em 30 de junho (faltam 12 dias)"
```

---

## 🔐 Segurança e Privacidade

```
┌─────────────────────────────┐
│    Seu Celular (WhatsApp)   │
└──────────────┬──────────────┘
               │
       🔒 Criptografia End-to-End
               │
┌──────────────▼──────────────┐
│   WhatsMeow Protocol         │
│   (Autenticação local)       │
└──────────────┬──────────────┘
               │
       🔒 Sessão persistida
       (arquivo: store.db)
               │
┌──────────────▼──────────────┐
│   Servidor Local             │
│   (Seu Computador)          │
│   • Dados locais             │
│   • Processamento local      │
│   • Nenhum upload externo   │
└─────────────────────────────┘

✅ Todos os dados ficam em seu computador
✅ Você controla sua sessão
✅ Sem compartilhamento com terceiros (MVP)
```

---

## 📈 Performance

### Latência End-to-End

```
Você escreve mensagem: 0ms
         ↓ (transmissão WhatsApp)
Servidor recebe: ~500ms
         ↓
Processa com IA: ~2s (Qwen local)
         ↓
Salva em banco: ~100ms
         ↓
Envia confirmação: ~500ms
         ↓
Você recebe: ~1s
         ├────────────────────────┤
         Total: ~4-5 segundos
```

### Escalabilidade

```
Workers paralelos com Redis:

1. Message chega
   ↓
2. Enfileira Redis
   ↓ (paralelo)
3. Worker AI processa
4. Worker Vision processa
5. Worker Email processa
   ↓
6. Todos salvam simultaneamente

Resultado: Múltiplas mensagens em paralelo!
```

---

## 🚀 Próximas Integrações

### Fase 2: Email + OCR

```
Email chega
    ↓
Gmail API captura
    ↓
Tesseract OCR extrai texto
    ↓
Vision AI analisa imagens
    ↓
"Anexei edital de concurso para você analisar"
    ↓
Cria task automaticamente
```

### Fase 3: Dashboard

```
Browser (React)
    ↓
GET /api/tasks
GET /api/analytics
GET /api/timeline
    ↓
Visualiza tudo em tempo real:
• Gráficos de carga
• Timeline de tarefas
• Estatísticas pessoais
```

### Fase 4: Integração Calendário

```
Seu calendário (Google Calendar / Outlook)
    ↓
Sync automático com lembretes
    ↓
Blocos de tempo:
• Evita conflitos
• Sugere slots ótimos
• Convida automaticamente
```

---

## 💡 Casos de Uso Reais

### Case 1: Concurso Público
```
"Saiu edital do concurso que quero!"
    ↓
Sistema detecta + armazena
    ↓
Progressivamente notifica sobre prazos
    ↓
Você não perde a inscrição ✅
```

### Case 2: Múltiplos Projetos
```
Você: "Deploy segunda as 10h, reunião terça 14h, relatório quinta"
    ↓
Sistema separa em 3 tarefas
    ↓
Agenda organizada automaticamente
    ↓
Você foca, não gerencia ✅
```

### Case 3: Contexto com Histórico
```
Você: "Preciso estudar para a prova"
    ↓
Sistema lembra: "Você mencionou prova de quarta agora"
    ↓
Relaciona com outras tarefas de estudo
    ↓
Cria blocos de tempo para preparação
    ↓
Você está preparado ✅
```

---

## 📞 Resumo

| Aspecto | MVP (Agora) | Futuro |
|--------|-----------|--------|
| **Captura** | WhatsApp | Email, PDFs, Voice |
| **AI** | Regex simples | Qwen local + Vision |
| **Armazena** | Memória RAM | PostgreSQL + Vetores |
| **Busca** | Lista simples | Semântica inteligente |
| **Notif** | Console | WhatsApp privado |
| **Dashboard** | Nenhum | React completo |

**Resultado Final:** Um segundo cérebro digital que não deixa você esquecer nada! 🧠✨
