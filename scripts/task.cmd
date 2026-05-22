@echo off
setlocal

if "%~1"=="" goto :usage

set "action=%~1"
if /I "%action%"=="run" (
    shift
    goto :run
)

echo Acao desconhecida: %action%
goto :usage

:run
if "%~1"=="" (
    echo Informe o caminho do script PowerShell.
    goto :usage
)

set "script=%~1"
if not exist "%script%" (
    echo Script nao encontrado: %script%
    exit /b 1
)

powershell -ExecutionPolicy Bypass -File "%script%"
exit /b %errorlevel%

:usage
echo Uso: task run caminho-do-script.ps1
exit /b 1