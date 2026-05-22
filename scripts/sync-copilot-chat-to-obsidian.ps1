param(
    [string]$VaultPath = "",
    [string]$WorkspaceStorageRoot = "",
    [string]$SessionId = ""
)

$ErrorActionPreference = "Stop"
$projectRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))

if ([string]::IsNullOrWhiteSpace($VaultPath)) {
    $VaultPath = Join-Path $projectRoot "second-brain"
} elseif (-not [System.IO.Path]::IsPathRooted($VaultPath)) {
    $VaultPath = Join-Path $projectRoot $VaultPath
}

if ([string]::IsNullOrWhiteSpace($WorkspaceStorageRoot)) {
    $WorkspaceStorageRoot = Join-Path $env:APPDATA "Code\\User\\workspaceStorage"
}

function Ensure-Directory {
    param([string]$Path)
    if (-not (Test-Path -Path $Path)) {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }
}

function Get-LatestTranscript {
    param([string]$Root)

    if (-not (Test-Path -Path $Root)) {
        throw "Workspace storage nao encontrado: $Root"
    }

    $file = Get-ChildItem -Path $Root -Recurse -File -Filter *.jsonl -ErrorAction SilentlyContinue |
        Where-Object { $_.FullName -match 'GitHub\.copilot-chat\\transcripts\\[^\\]+\.jsonl$' } |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1

    if (-not $file) {
        throw "Nenhum transcript do Copilot encontrado em $Root"
    }

    return $file.FullName
}

function Get-TranscriptBySession {
    param(
        [string]$Root,
        [string]$Session
    )

    $file = Get-ChildItem -Path $Root -Recurse -File -Filter "$Session.jsonl" -ErrorAction SilentlyContinue |
        Where-Object { $_.FullName -match 'GitHub\.copilot-chat\\transcripts\\[^\\]+\.jsonl$' } |
        Select-Object -First 1

    if (-not $file) {
        throw "Transcript da sessao $Session nao encontrado em $Root"
    }

    return $file.FullName
}

function Extract-SessionId {
    param([string]$TranscriptPath)
    return [System.IO.Path]::GetFileNameWithoutExtension($TranscriptPath)
}

function Build-MarkdownFromTranscript {
    param([string]$TranscriptPath)

    $lines = Get-Content -Path $TranscriptPath -Encoding UTF8
    $entries = New-Object System.Collections.Generic.List[string]

    foreach ($line in $lines) {
        if ([string]::IsNullOrWhiteSpace($line)) { continue }

        try {
            $event = $line | ConvertFrom-Json
        } catch {
            continue
        }

        $type = [string]$event.type
        $ts = [string]$event.timestamp

        if ($type -eq "user.message") {
            $content = [string]$event.data.content
            if (-not [string]::IsNullOrWhiteSpace($content)) {
                $entries.Add("### User ($ts)")
                $entries.Add("")
                $entries.Add($content)
                $entries.Add("")
            }
        } elseif ($type -eq "assistant.message") {
            $content = [string]$event.data.content
            if (-not [string]::IsNullOrWhiteSpace($content)) {
                $entries.Add("### Assistant ($ts)")
                $entries.Add("")
                $entries.Add($content)
                $entries.Add("")
            }
        }
    }

    return $entries
}

# Resolve transcript source
$transcriptPath = ""
if ([string]::IsNullOrWhiteSpace($SessionId)) {
    $transcriptPath = Get-LatestTranscript -Root $WorkspaceStorageRoot
} else {
    $transcriptPath = Get-TranscriptBySession -Root $WorkspaceStorageRoot -Session $SessionId
}

$session = Extract-SessionId -TranscriptPath $transcriptPath
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# Ensure vault structure
$inboxPath = Join-Path $VaultPath "00-Inbox"
$archivePath = Join-Path $VaultPath "99-Archive\\copilot-transcripts"
$indexPath = Join-Path $inboxPath "Chat Index.md"
$chatNotePath = Join-Path $inboxPath ("Chat - " + $session + ".md")

Ensure-Directory -Path $VaultPath
Ensure-Directory -Path $inboxPath
Ensure-Directory -Path $archivePath

# Archive raw transcript
$rawArchivePath = Join-Path $archivePath ("{0}-{1}.jsonl" -f (Get-Date -Format "yyyyMMdd-HHmmss"), $session)
Copy-Item -Path $transcriptPath -Destination $rawArchivePath -Force

# Build markdown chat note
$entries = Build-MarkdownFromTranscript -TranscriptPath $transcriptPath

$header = @(
    "# Chat Session $session",
    "",
    "- Synced at: $timestamp",
    "- Source: $transcriptPath",
    "- Raw archive: $rawArchivePath",
    "",
    "## Messages",
    ""
)

$all = $header + $entries
$all | Set-Content -Path $chatNotePath -Encoding UTF8

# Update index note
if (-not (Test-Path -Path $indexPath)) {
    "# Chat Index`n`n" | Set-Content -Path $indexPath -Encoding UTF8
}

$indexContent = Get-Content -Path $indexPath -Encoding UTF8
$linkLine = "- [[Chat - $session]] - $timestamp"
if ($indexContent -notcontains $linkLine) {
    Add-Content -Path $indexPath -Value $linkLine -Encoding UTF8
}

Write-Host "Transcript synced successfully."
Write-Host "Session note: $chatNotePath"
Write-Host "Raw archive: $rawArchivePath"
