#!/usr/bin/env bash
# ==============================================================================
# Self-Signed TLS Certificate Generator for ERP Staging & Local Testing
# ==============================================================================
set -euo pipefail

SSL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOMAIN="${1:-erp.example.sa}"
DAYS=365

echo "==> Generating self-signed SSL/TLS certificate for: ${DOMAIN} in ${SSL_DIR}"

openssl req -x509 -nodes -days "${DAYS}" -newkey rsa:2048 \
    -keyout "${SSL_DIR}/erp.key" \
    -out "${SSL_DIR}/erp.crt" \
    -subj "/C=SA/ST=Riyadh/L=Riyadh/O=Saudi Enterprise ERP/OU=Operations/CN=${DOMAIN}" \
    -addext "subjectAltName=DNS:${DOMAIN},DNS:localhost,IP:127.0.0.1"

chmod 600 "${SSL_DIR}/erp.key"
chmod 644 "${SSL_DIR}/erp.crt"

echo "==> Successfully generated ${SSL_DIR}/erp.crt and ${SSL_DIR}/erp.key"
