<#
.SYNOPSIS
    Saudi Enterprise ERP - System Doctor Diagnostic Utility (PowerShell)
.DESCRIPTION
    Validates runtime tools, database migration files, Next.js frontend,
    production configuration, and regulatory settings for the Saudi ERP platform.
#>

[CmdletBinding()]
param(
    [string]$ApiBaseUrl = "http://localhost:8080",
    [string]$ProductionEnvFile = "",
    [switch]$RequireApi
)

Set-StrictMode -Version Latest

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

$WorkspaceRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
if (-not $ProductionEnvFile) {
    $ProductionEnvFile = Join-Path $WorkspaceRoot "deployments/compose/.env.production"
    if (-not (Test-Path $ProductionEnvFile)) {
        $ProductionEnvFile = Join-Path $WorkspaceRoot ".env.production"
    }
}

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
    $RequiredGo = (Select-String -Path (Join-Path $WorkspaceRoot "go.mod") -Pattern '^go\s+([0-9.]+)$').Matches.Groups[1].Value
    $InstalledGo = $GoVer.TrimStart("go")
    if ([version]$InstalledGo -lt [version]$RequiredGo) {
        Write-Fail "Go $RequiredGo or newer is required; found $InstalledGo"
    } else {
        Write-Pass "Go runtime satisfies go.mod ($InstalledGo >= $RequiredGo)"
    }
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
    docker info *> $null
    if ($LASTEXITCODE -eq 0) {
        Write-Pass "Docker engine is reachable: $DockerVer"
    } else {
        Write-Warn "Docker CLI is installed but the engine is not reachable"
    }
} else {
    Write-Warn "Docker engine not found in PATH"
}

# 2. Check Regulatory & Timezone Settings
Write-Host ""
Write-Host "==> 2. Checking Regulatory & Timezone Settings..." -ForegroundColor Cyan
$CurrentTz = [System.TimeZoneInfo]::Local.Id
$CurrentOffset = [System.TimeZoneInfo]::Local.BaseUtcOffset.Hours
if ($CurrentOffset -eq 3 -or $CurrentTz -match "Arab|Riyadh") {
    Write-Pass "Host timezone is Asia/Riyadh-compatible ($CurrentTz, UTC+3)"
} else {
    Write-Warn "Host timezone is '$CurrentTz' (UTC$CurrentOffset). Production containers use Asia/Riyadh; host timezone alone does not establish regulatory compliance."
}

# 3. Check Database Migrations & Schemas
Write-Host ""
Write-Host "==> 3. Checking Database Migrations & Schemas..." -ForegroundColor Cyan
$MigrationDir = Join-Path $WorkspaceRoot "database/migrations"
if (Test-Path $MigrationDir) {
    $MigCount = (Get-ChildItem -Path $MigrationDir -Filter "*.sql").Count
    if ($MigCount -gt 0) {
        Write-Pass "Found $MigCount SQL migration scripts in database/migrations"
        $MigrationVersions = Get-ChildItem -Path $MigrationDir -Filter "*.sql" | ForEach-Object {
            if ($_.Name -match '^(\d{6})_') { $Matches[1] } else { Write-Fail "Invalid migration filename: $($_.Name)" }
        }
        $Duplicates = $MigrationVersions | Group-Object | Where-Object Count -gt 1
        if ($Duplicates) {
            Write-Fail "Duplicate migration versions: $($Duplicates.Name -join ', ')"
        } else {
            Write-Pass "Migration versions are unique; latest is $($MigrationVersions | Sort-Object | Select-Object -Last 1)"
        }
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
        Write-Warn "node_modules missing in apps/web. Run 'npm ci' in apps/web."
    }
    if (Test-Path (Join-Path $WebDir "package-lock.json")) {
        Write-Pass "Frontend dependency lockfile is present"
    } else {
        Write-Fail "apps/web/package-lock.json is required for reproducible npm ci builds"
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

$DevCompose = Join-Path $WorkspaceRoot "docker-compose.yml"
if (Get-Command docker -ErrorAction SilentlyContinue) {
    docker compose -f $DevCompose config --quiet
    if ($LASTEXITCODE -eq 0) {
        Write-Pass "Development Compose renders successfully"
    } else {
        Write-Fail "Development Compose configuration is invalid"
    }
}

$EnvExample = Join-Path $WorkspaceRoot ".env.production.example"
if (Test-Path $EnvExample) {
    $RequiredVariables = @(
        "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "DATABASE_URL", "MIGRATION_DATABASE_URL", "APP_URL",
        "ERP_MIGRATOR_PASSWORD", "ERP_APP_PASSWORD", "ERP_BACKUP_PASSWORD", "BACKUP_ENCRYPTION_KEY",
        "TLS_CERT_PATH", "TLS_KEY_PATH", "MFA_ENCRYPTION_KEY", "METRICS_BEARER_TOKEN",
        "SMTP_HOST", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM"
    )
    $ExampleContent = Get-Content -Raw $EnvExample
    $MissingVariables = $RequiredVariables | Where-Object { $ExampleContent -notmatch "(?m)^$([regex]::Escape($_))=" }
    if ($MissingVariables) {
        Write-Fail "Production environment template misses: $($MissingVariables -join ', ')"
    } else {
        Write-Pass "Production environment template contains all required variables"
    }
} else {
    Write-Fail ".env.production.example template missing at repository root"
}

$ProductionEnv = $ProductionEnvFile
if (Test-Path $ProductionEnv) {
    if (Get-Command docker -ErrorAction SilentlyContinue) {
        docker compose --env-file $ProductionEnv -f $ComposeFile config --quiet
        if ($LASTEXITCODE -eq 0) {
            Write-Pass "Production Compose renders with .env.production"
        } else {
            Write-Fail "Production Compose or .env.production is invalid"
        }
    }
    if (Select-String -Path $ProductionEnv -Quiet -Pattern 'replace-with|changeme|example\.(com|sa)|Generate64Byte') {
        Write-Fail ".env.production still contains placeholder values"
    }
} else {
    Write-Warn "No active .env.production found; production Compose secret values were not validated"
}

# 6. Service Health Check (if API is running locally)
Write-Host ""
Write-Host "==> 6. Probing Running Endpoints (Local)..." -ForegroundColor Cyan
try {
    $Response = Invoke-WebRequest -Uri "$ApiBaseUrl/healthz" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
    if ($Response.StatusCode -eq 200) {
        Write-Pass "API liveness probe passed at $ApiBaseUrl/healthz"
        try {
            $Ready = Invoke-WebRequest -Uri "$ApiBaseUrl/readyz" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
            if ($Ready.StatusCode -eq 200) { Write-Pass "API readiness probe passed" } else { Write-Fail "API readiness probe returned $($Ready.StatusCode)" }
        } catch {
            Write-Fail "API is alive but readiness failed at $ApiBaseUrl/readyz; inspect database schema and document storage"
        }
    } else {
        Write-Fail "API liveness responded with status $($Response.StatusCode)"
    }
} catch {
    if ($RequireApi -or $_.Exception.PSObject.Properties['Response']) {
        Write-Fail "API liveness failed at $ApiBaseUrl/healthz"
    } else {
        Write-Warn "API is not currently reachable at $ApiBaseUrl (offline checks still completed; use -RequireApi for deployment validation)"
    }
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
