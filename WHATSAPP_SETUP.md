# WhatsApp Integration Guide - WhatsMeow

## Como Funciona

O sistema usa a biblioteca **WhatsMeow** que se conecta ao WhatsApp Web através do protocolo reverso engineered. Quando ativo, a aplicação:

1. **Escaneia um QR Code** - Você lê com seu celular
2. **Autentica sua sessão** - Fica sincronizado com seu número
3. **Monitora mensagens** - Em tempo real (24/7 se o servidor estiver rodando)
4. **Extrai compromissos** - Analisa tudo que você escreve
5. **Cria lembretes** - Automático e inteligente

## Requisitos

- Seu número pessoal do WhatsApp
- Aplicativo WhatsApp instalado no celular
- Go 1.22+

## Passo 1: Instalar Dependências

```bash
go mod tidy
```

Isso vai baixar:
- `go.mau.fi/whatsmeow` - Cliente WhatsApp
- `google.golang.org/protobuf` - Serialização de mensagens

## Passo 2: Rodar a Aplicação

```bash
go run ./cmd/main.go
```

Vai aparecer algo como:

```
Assistente Cognitivo Pessoal iniciado.
Módulo WhatsApp iniciando (WhatsMeow)...
[QR] Escaneie este QR Code com seu WhatsApp:

████████████████████████████████
████████████████████████████████
████████████████████████████████
...
```

## Passo 3: Escanear QR Code

1. Abra o WhatsApp no seu celular
2. Vá para: **Configurações → Celular conectado/Dispositivos conectados**
3. Toque em **"Conectar um dispositivo"**
4. Aponte a câmera para o QR Code que apareceu no terminal

## Passo 4: Pronto! Agora Envie uma Mensagem

Escreva uma mensagem em qualquer chat do WhatsApp:

```
"Sexta tenho entrevista às 14h"
```

Você receberá:

```
✅ Tarefa criada!
📌 Sexta tenho entrevista às 14h
⏰ 22/05 14:00
```

## Recursos Detectados Automaticamente

| Mensagem | Detecta |
|----------|---------|
| "Sexta às 14h" | Data: sexta, Hora: 14h |
| "Amanhã reunião" | Data: amanhã, Tipo: reunião |
| "15 de julho prova" | Data exata: 15/07 |
| "Em 3 dias visita" | Data relativa |
| "18:30 dentista" | Hora específica |

## Endpoints da API

Se preferir usar via HTTP (sem WhatsApp conectado):

### Enviar mensagem manualmente
```bash
curl -X POST http://localhost:8080/ingest/whatsapp \
  -H "Content-Type: application/json" \
  -d '{"message":"sexta tenho prova as 19h"}'
```

### Ver tarefas criadas
```bash
curl http://localhost:8080/tasks
```

### Status de conexão
```bash
curl http://localhost:8080/status
```

Retorna:
```json
{
  "connected": true,
  "client": true
}
```

## Arquivos Gerados

A aplicação cria um diretório:

```
whatsapp_data/
├── store.db          # Sessão persistida (não é preciso fazer login novamente)
```

Se quiser fazer logout, delete este arquivo e rode novamente para escanear novo QR.

## Troubleshooting

| Problema | Solução |
|----------|---------|
| **QR Code não aparece** | Seu número pode estar banido. Use outro WhatsApp. |
| **"Desconectado" após 1h** | Normal. WhatsMeow reconecta automaticamente. |
| **Mensagens não chegam** | Cheque se o servidor está rodando em background. |
| **Muitas mensagens sendo processadas** | Adicione filtro para palavras-chave. |

## Segurança

⚠️ **Importante:**
- Este sistema acessa TODAS as mensagens que chegam
- Não compartilhe o banco `store.db` - é sua sessão pessoal
- Se notar acesso estranho, delete `whatsapp_data/` e faça login novamente

## Próximos Passos

1. Adicionar filtro de palavras-chave (não processar bots/notificações)
2. Criar modo "modo silencioso" (não responder automaticamente)
3. Sincronizar com Gmail/Calendário
4. Dashboard em tempo real das tarefas

## Links Úteis

- [WhatsMeow Docs](https://pkg.go.dev/go.mau.fi/whatsmeow)
- [WhatsApp Protocol](https://github.com/tulir/whatsmeow)
