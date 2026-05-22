# Project Second Brain (Obsidian)

Este vault foi criado para manter contexto permanente do projeto no Obsidian.

## Estrutura

- `00-Inbox/` Capturas rapidas e logs de chat
- `01-Daily/` Notas diarias
- `02-Decisions/` ADRs e decisoes tecnicas
- `03-Architecture/` Diagramas e visao de arquitetura
- `04-Tasks/` Planos e backlog
- `05-Incidents/` Falhas e pos-mortem
- `99-Archive/copilot-transcripts/` Transcripts brutos do Copilot

## Sincronizar chat do Copilot

Use o script abaixo na raiz do projeto:

```powershell
./scripts/sync-copilot-chat-to-obsidian.ps1
```

Opcoes uteis:

```powershell
# sincroniza sessao especifica
./scripts/sync-copilot-chat-to-obsidian.ps1 -SessionId "3a17e94c-a2ec-44e0-a724-c7e888120a51"

# muda caminho do vault
./scripts/sync-copilot-chat-to-obsidian.ps1 -VaultPath ".\\second-brain"
```

## Fluxo recomendado

1. Rode o script ao final de cada bloco de trabalho.
2. Revise o arquivo em `00-Inbox/`.
3. Promova insights para `02-Decisions/` e `03-Architecture/`.
