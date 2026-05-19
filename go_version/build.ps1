$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
go build -o daily-tasks.exe ./cmd/daily-tasks
Write-Host "Built: $(Join-Path $PSScriptRoot 'daily-tasks.exe')"
