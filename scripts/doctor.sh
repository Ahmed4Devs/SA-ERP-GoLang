#!/usr/bin/env bash
# ==============================================================================
# Saudi Enterprise ERP - System Doctor Diagnostic Utility (POSIX Bash)
# Checks environment, database migrations, security secrets, and runtime health
# ==============================================================================
set -u

WORKSPACE_ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PASSED=0
WARNINGS=0
FAILURES=0

print_header() {
    echo -e "${BLUE}====================================================================${NC}"
    echo -e "${BLUE}        SAUDI ENTERPRISE ERP - SYSTEM DOCTOR DIAGNOSTIC TOOL        ${NC}"
    echo -e "${BLUE}====================================================================${NC}"
    echo -e "Timestamp: $(date +"%Y-%m-%d %H:%M:%S")"
    echo ""
}

check_pass() {
    echo -e " [${GREEN}PASS${NC}] $1"
    PASSED=$((PASSED + 1))
}

check_warn() {
    echo -e " [${YELLOW}WARN${NC}] $1"
    WARNINGS=$((WARNINGS + 1))
}

check_fail() {
    echo -e " [${RED}FAIL${NC}] $1"
    FAILURES=$((FAILURES + 1))
}

print_header

# 1. Check Tooling & Runtimes
echo "==> 1. Checking Tooling & Runtimes..."
if command -v go &> /dev/null; then
    GO_VER=$(go version | awk '{print $3}')
    REQUIRED_GO=$(awk '$1 == "go" {print $2; exit}' "${WORKSPACE_ROOT}/go.mod")
    INSTALLED_GO=${GO_VER#go}
    if [ "$(printf '%s\n' "${REQUIRED_GO}" "${INSTALLED_GO}" | sort -V | head -n1)" = "${REQUIRED_GO}" ]; then
        check_pass "Go runtime satisfies go.mod (${INSTALLED_GO} >= ${REQUIRED_GO})"
    else
        check_fail "Go ${REQUIRED_GO} or newer is required; found ${INSTALLED_GO}"
    fi
else
    check_fail "Go runtime not found in PATH"
fi

if command -v node &> /dev/null; then
    NODE_VER=$(node -v)
    check_pass "Node.js found: ${NODE_VER}"
else
    check_warn "Node.js not found in PATH (required for Next.js web frontend)"
fi

if command -v docker &> /dev/null; then
    DOCKER_VER=$(docker --version | awk '{print $3}' | tr -d ',')
    if docker info >/dev/null 2>&1; then
        check_pass "Docker engine is reachable: ${DOCKER_VER}"
    else
        check_warn "Docker CLI is installed but the engine is not reachable"
    fi
else
    check_warn "Docker not found in PATH"
fi

# 2. Check Timezone & Saudi Compliance
echo ""
echo "==> 2. Checking Regulatory & Timezone Settings..."
CURRENT_TZ="${TZ:-$(cat /etc/timezone 2>/dev/null || date +%Z)}"
if [[ "${CURRENT_TZ}" =~ (Riyadh|AST|Asia/Riyadh|\+03) ]]; then
    check_pass "Timezone matches Saudi Arabia regulatory standard (${CURRENT_TZ})"
else
    check_warn "Timezone is '${CURRENT_TZ}'. Recommended: 'Asia/Riyadh' (UTC+3) for ZATCA compliance."
fi

# 3. Check Database Migrations & Structure
echo ""
echo "==> 3. Checking Database Migrations & Schemas..."
MIGRATION_DIR="${WORKSPACE_ROOT}/database/migrations"
if [ -d "${MIGRATION_DIR}" ]; then
    MIG_COUNT=$(find "${MIGRATION_DIR}" -type f -name "*.sql" | wc -l)
    if [ "${MIG_COUNT}" -gt 0 ]; then
        check_pass "Found ${MIG_COUNT} SQL migrations in database/migrations"
        INVALID_MIGRATIONS=$(find "${MIGRATION_DIR}" -maxdepth 1 -type f -name '*.sql' -printf '%f\n' | grep -Ev '^[0-9]{6}_.+\.sql$' || true)
        DUPLICATE_MIGRATIONS=$(find "${MIGRATION_DIR}" -maxdepth 1 -type f -name '*.sql' -printf '%f\n' | cut -c1-6 | sort | uniq -d)
        if [ -n "${INVALID_MIGRATIONS}" ]; then
            check_fail "Migration filenames must start with six digits: ${INVALID_MIGRATIONS}"
        elif [ -n "${DUPLICATE_MIGRATIONS}" ]; then
            check_fail "Duplicate migration versions: ${DUPLICATE_MIGRATIONS}"
        else
            LATEST_MIGRATION=$(find "${MIGRATION_DIR}" -maxdepth 1 -type f -name '*.sql' -printf '%f\n' | sort | tail -n1)
            check_pass "Migration versions are unique; latest is ${LATEST_MIGRATION%%_*}"
        fi
    else
        check_fail "Migration directory database/migrations exists but contains no .sql files"
    fi
else
    check_fail "Migrations directory not found at ${MIGRATION_DIR}"
fi

# 4. Check Frontend Web App
echo ""
echo "==> 4. Checking Next.js Web Frontend..."
WEB_DIR="${WORKSPACE_ROOT}/apps/web"
if [ -d "${WEB_DIR}" ]; then
    if [ -f "${WEB_DIR}/next.config.ts" ] || [ -f "${WEB_DIR}/next.config.js" ] || [ -f "${WEB_DIR}/next.config.mjs" ]; then
        check_pass "Next.js configuration found in apps/web"
    else
        check_fail "next.config.* missing in apps/web"
    fi
    if [ -d "${WEB_DIR}/node_modules" ]; then
        check_pass "Frontend node_modules present"
    else
        check_warn "node_modules missing in apps/web. Run 'npm install' inside apps/web."
    fi
    if [ -f "${WEB_DIR}/package-lock.json" ]; then
        check_pass "Frontend dependency lockfile is present"
    else
        check_fail "apps/web/package-lock.json is required for reproducible npm ci builds"
    fi
else
    check_fail "Frontend directory apps/web not found"
fi

# 5. Check Production Deployments & Configuration
echo ""
echo "==> 5. Checking Deployments & Configuration..."
COMPOSE_FILE="${WORKSPACE_ROOT}/deployments/compose/docker-compose.prod.yml"
if [ -f "${COMPOSE_FILE}" ]; then
    check_pass "Production compose configuration found at deployments/compose/docker-compose.prod.yml"
else
    check_fail "Production compose file missing: ${COMPOSE_FILE}"
fi

if command -v docker &> /dev/null; then
    if docker compose -f "${WORKSPACE_ROOT}/docker-compose.yml" config --quiet; then
        check_pass "Development Compose renders successfully"
    else
        check_fail "Development Compose configuration is invalid"
    fi
fi

ENV_EXAMPLE="${WORKSPACE_ROOT}/.env.production.example"
if [ -f "${ENV_EXAMPLE}" ]; then
    MISSING_KEYS=""
    for key in POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB DATABASE_URL APP_URL TLS_CERT_PATH TLS_KEY_PATH MFA_ENCRYPTION_KEY METRICS_BEARER_TOKEN SMTP_HOST SMTP_USERNAME SMTP_PASSWORD SMTP_FROM; do
        if ! grep -q "^${key}=" "${ENV_EXAMPLE}"; then MISSING_KEYS="${MISSING_KEYS} ${key}"; fi
    done
    if [ -n "${MISSING_KEYS}" ]; then
        check_fail "Production environment template misses:${MISSING_KEYS}"
    else
        check_pass "Production environment template contains all required variables"
    fi
else
    check_fail ".env.production.example template missing at repository root"
fi

ENV_PROD="${WORKSPACE_ROOT}/.env.production"
if [ -f "${ENV_PROD}" ]; then
    check_pass "Found active .env.production"
    if command -v docker &> /dev/null; then
        if docker compose --env-file "${ENV_PROD}" -f "${COMPOSE_FILE}" config --quiet; then
            check_pass "Production Compose renders with .env.production"
        else
            check_fail "Production Compose or .env.production is invalid"
        fi
    fi
    if grep -Eqi 'replace-with|changeme|example\.(com|sa)|Generate64Byte' "${ENV_PROD}"; then
        check_fail ".env.production still contains placeholder values"
    fi
else
    check_warn "No active .env.production found; production Compose secret values were not validated"
fi

# 6. Service Health Check (if services are listening locally)
echo ""
echo "==> 6. Probing Running Endpoints (Local)..."
if command -v curl &> /dev/null; then
    API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${API_BASE_URL}/healthz" 2>/dev/null || echo "DOWN")
    if [ "${API_STATUS}" == "200" ]; then
        check_pass "API liveness probe passed at ${API_BASE_URL}/healthz"
        READY_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${API_BASE_URL}/readyz" 2>/dev/null || echo "DOWN")
        if [ "${READY_STATUS}" == "200" ]; then check_pass "API readiness probe passed"; else check_fail "API readiness probe failed (${READY_STATUS})"; fi
    else
        check_warn "API is not reachable at ${API_BASE_URL} (Status: ${API_STATUS})"
    fi
else
    check_warn "curl not available to test local health endpoints"
fi

# Final Summary
echo ""
echo -e "${BLUE}====================================================================${NC}"
echo -e "Doctor Summary: ${GREEN}${PASSED} Passed${NC} | ${YELLOW}${WARNINGS} Warnings${NC} | ${RED}${FAILURES} Failures${NC}"
echo -e "${BLUE}====================================================================${NC}"

if [ "${FAILURES}" -gt 0 ]; then
    echo -e "${RED}Result: System has critical issues that need resolution.${NC}"
    exit 1
else
    echo -e "${GREEN}Result: System baseline checks passed!${NC}"
    exit 0
fi
