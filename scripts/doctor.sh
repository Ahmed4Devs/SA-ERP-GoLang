#!/usr/bin/env bash
# ==============================================================================
# Saudi Enterprise ERP - System Doctor Diagnostic Utility (POSIX Bash)
# Checks environment, database migrations, security secrets, and runtime health
# ==============================================================================
set -u

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
    check_pass "Go runtime found: ${GO_VER}"
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
    check_pass "Docker engine found: ${DOCKER_VER}"
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
MIGRATION_DIR="$(dirname "$0")/../database/migrations"
if [ -d "${MIGRATION_DIR}" ]; then
    MIG_COUNT=$(find "${MIGRATION_DIR}" -type f -name "*.sql" | wc -l)
    if [ "${MIG_COUNT}" -gt 0 ]; then
        check_pass "Found ${MIG_COUNT} SQL migrations in database/migrations"
    else
        check_fail "Migration directory database/migrations exists but contains no .sql files"
    fi
else
    check_fail "Migrations directory not found at ${MIGRATION_DIR}"
fi

# 4. Check Frontend Web App
echo ""
echo "==> 4. Checking Next.js Web Frontend..."
WEB_DIR="$(dirname "$0")/../apps/web"
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
else
    check_fail "Frontend directory apps/web not found"
fi

# 5. Check Production Deployments & Configuration
echo ""
echo "==> 5. Checking Deployments & Configuration..."
COMPOSE_FILE="$(dirname "$0")/../deployments/compose/docker-compose.prod.yml"
if [ -f "${COMPOSE_FILE}" ]; then
    check_pass "Production compose configuration found at deployments/compose/docker-compose.prod.yml"
else
    check_fail "Production compose file missing: ${COMPOSE_FILE}"
fi

ENV_PROD="$(dirname "$0")/../deployments/compose/.env.production"
if [ -f "${ENV_PROD}" ]; then
    check_pass "Found active .env.production"
    # Detect default placeholder secrets
    if grep -q "ChangeMeInProduction" "${ENV_PROD}" || grep -q "Generate64Byte" "${ENV_PROD}"; then
        check_warn ".env.production contains placeholder secrets! Replace before launching in production."
    fi
else
    check_warn "No .env.production file found in deployments/compose. (Copy from .env.production.example)"
fi

# 6. Service Health Check (if services are listening locally)
echo ""
echo "==> 6. Probing Running Endpoints (Local)..."
if command -v curl &> /dev/null; then
    API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/healthz 2>/dev/null || echo "DOWN")
    if [ "${API_STATUS}" == "200" ]; then
        check_pass "Local API server responded healthy (HTTP 200 at :8080/healthz)"
    else
        check_warn "Local API server not responding on :8080/healthz (Status: ${API_STATUS})"
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
