<#
.SYNOPSIS
    Saudi Enterprise ERP - System Doctor Diagnostic Utility (PowerShell)
.DESCRIPTION
    Validates runtime tools, database migration files, Next.js frontend,
    production configuration, and regulatory settings for the Saudi ERP platform.
#>

[CmdletBinding()]
param()

$Passed = 0
$Warnings = 0
$Failures = 0

function Write-Pass {
    param([string]$Message)
    Write-Host " [PASS] $Message" -ForegroundColor Green
    $script:Passed++
}

function Write-Warn {
    param([string]$Message)
    Write-Host " [WARN] $Message" -ForegroundColor Yellow
    $script:Warnings++
}

function Write-Fail {
    param([string]$Message)
    Write-Host " [FAIL] $Message" -ForegroundColor Red
    $script:Failures++
}

$WorkspaceRoot = Resolve-Path (Join-Path $PSScriptRoot "..")

Write-Host "====================================================================" -ForegroundColor Cyan
Write-Host "        SAUDI ENTERPRISE ERP - SYSTEM DOCTOR DIAGNOSTIC TOOL        " -ForegroundColor Cyan
Write-Host "====================================================================" -ForegroundColor Cyan
Write-Host "Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
Write-Host "Workspace: $WorkspaceRoot"
Write-Host ""

# 1. Check Tooling & Runtimes
Write-Host "==> 1. Checking Tooling & Runtimes..." -ForegroundColor Cyan
if (Get-Command go -ErrorAction SilentlyContinue) {
    $GoVer = (go version).Split(" ")[2]
    Write-Pass "Go runtime found: $GoVer"
} else {
    Write-Fail "Go runtime not found in PATH"
}

if (Get-Command node -ErrorAction SilentlyContinue) {
    $NodeVer = (node -v)
    Write-Pass "Node.js runtime found: $NodeVer"
} else {
    Write-Warn "Node.js not found in PATH (required for web frontend)"
}

if (Get-Command docker -ErrorAction SilentlyContinue) {
    $DockerVer = (docker --version)
    Write-Pass "Docker found: $DockerVer"
} else {
    Write-Warn "Docker engine not found in PATH"
}

# 2. Check Regulatory & Timezone Settings
Write-Host ""
Write-Host "==> 2. Checking Regulatory & Timezone Settings..." -ForegroundColor Cyan
$CurrentTz = [System.TimeZoneInfo]::Local.Id
$CurrentOffset = [System.TimeZoneInfo]::Local.BaseUtcOffset.Hours
if ($CurrentOffset -eq 3 -or $CurrentTz -match "Arab|Riyadh") {
    Write-Pass "Timezone is aligned with Saudi Arabia regulatory standard ($CurrentTz, UTC+3)"
} else {
    Write-Warn "Timezone is '$CurrentTz' (UTC$CurrentOffset). Recommended for production: 'Asia/Riyadh' (UTC+3) for ZATCA e-invoicing compliance."
}

# 3. Check Database Migrations & Schemas
Write-Host ""
Write-Host "==> 3. Checking Database Migrations & Schemas..." -ForegroundColor Cyan
$MigrationDir = Join-Path $WorkspaceRoot "database/migrations"
if (Test-Path $MigrationDir) {
    $MigCount = (Get-ChildItem -Path $MigrationDir -Filter "*.sql").Count
    if ($MigCount -gt 0) {
        Write-Pass "Found $MigCount SQL migration scripts in database/migrations"
    } else {
        Write-Fail "Migration directory database/migrations exists but contains no .sql files"
    }
} else {
    Write-Fail "Migration directory not found at $MigrationDir"
}

# 4. Check Frontend Web App
Write-Host ""
Write-Host "==> 4. Checking Next.js Web Frontend..." -ForegroundColor Cyan
$WebDir = Join-Path $WorkspaceRoot "apps/web"
if (Test-Path $WebDir) {
    $hasConfig = (Test-Path (Join-Path $WebDir "next.config.ts")) -or (Test-Path (Join-Path $WebDir "next.config.js")) -or (Test-Path (Join-Path $WebDir "next.config.mjs"))
    if ($hasConfig) {
        Write-Pass "Next.js configuration found in apps/web"
    } else {
        Write-Fail "next.config.* missing in apps/web"
    }
    if (Test-Path (Join-Path $WebDir "node_modules")) {
        Write-Pass "Frontend node_modules present"
    } else {
        Write-Warn "node_modules missing in apps/web. Run 'npm install' in apps/web."
    }
} else {
    Write-Fail "Frontend directory apps/web not found"
}

# 5. Check Production Deployments & Configuration
Write-Host ""
Write-Host "==> 5. Checking Deployments & Configuration..." -ForegroundColor Cyan
$ComposeFile = Join-Path $WorkspaceRoot "deployments/compose/docker-compose.prod.yml"
if (Test-Path $ComposeFile) {
    Write-Pass "Production compose configuration found at deployments/compose/docker-compose.prod.yml"
} else {
    Write-Fail "Production compose file missing: $ComposeFile"
}

$EnvExample = Join-Path $WorkspaceRoot "deployments/compose/.env.production.example"
if (Test-Path $EnvExample) {
    Write-Pass "Production environment example template found (.env.production.example)"
} else {
    Write-Fail ".env.production.example template missing in deployments/compose"
}

# 6. Service Health Check (if API is running locally)
Write-Host ""
Write-Host "==> 6. Probing Running Endpoints (Local)..." -ForegroundColor Cyan
try {
    $Response = Invoke-WebRequest -Uri "http://localhost:8080/healthz" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
    if ($Response.StatusCode -eq 200) {
        Write-Pass "Local API server responded healthy (HTTP 200 at :8080/healthz)"
    } else {
        Write-Warn "Local API responded with status $($Response.StatusCode)"
    }
} catch {
    Write-Warn "Local API server not currently listening on :8080/healthz (Service offline)"
}

# Final Summary
Write-Host ""
Write-Host "====================================================================" -ForegroundColor Cyan
Write-Host "Doctor Summary: $Passed Passed | $Warnings Warnings | $Failures Failures" -ForegroundColor $(if ($Failures -eq 0) { "Green" } else { "Red" })
Write-Host "====================================================================" -ForegroundColor Cyan

if ($Failures -gt 0) {
    Write-Host "Result: System has critical issues that require resolution." -ForegroundColor Red
    exit 1
} else {
    Write-Host "Result: System baseline checks passed successfully!" -ForegroundColor Green
    exit 0
}
