# Saudi Enterprise ERP - Production Operations Runbook

## 1. Overview & Architecture Topology
This document provides standard operating procedures (SOP) for the 24/7 operations, maintenance, and incident response of the Saudi Enterprise ERP production system.

### Container Topology
The production deployment orchestrates 7 primary containers via Docker Compose or Kubernetes:
- **`erp-nginx`**: Edge reverse proxy handling TLS termination, rate limiting, and HTTP/2 routing.
- **`erp-web`**: Next.js 15 standalone frontend web server (Port 3000).
- **`erp-api`**: Go core backend API service with GORM connection pooling (Port 8080).
- **`erp-worker`**: Go background processing worker for ZATCA clearance, PDF generation, and notifications.
- **`erp-migrate`**: Ephemeral schema migration runner executing Goose database migrations before API launch.
- **`erp-postgres`**: PostgreSQL 16 relational database with Saudi localized schema (`Asia/Riyadh` TZ).
- **`erp-redis`**: Redis 7 cache and high-throughput task queue broker.

---

## 2. Service Lifecycle Commands

### Starting the Production Stack
```bash
# Navigate to deployment directory
cd deployments/compose

# Validate environment configuration
cp .env.production.example .env.production
# Ensure secrets and credentials are fully set

# Pull & build containers with zero cache
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml build --no-cache

# Start stack in background
docker compose -f docker-compose.prod.yml up -d
```

### Checking Stack Status & Health
```bash
# View running status and container health
docker compose -f docker-compose.prod.yml ps

# Follow logs across all services
docker compose -f docker-compose.prod.yml logs -f --tail=100

# Inspect specific service logs (e.g. API or Worker)
docker compose -f docker-compose.prod.yml logs -f erp-api
docker compose -f docker-compose.prod.yml logs -f erp-worker
```

### Stopping & Graceful Restart
```bash
# Graceful restart of API service
docker compose -f docker-compose.prod.yml restart erp-api

# Stop full stack without data deletion
docker compose -f docker-compose.prod.yml down

# Graceful reload of Nginx configurations without dropping connections
docker compose -f docker-compose.prod.yml exec nginx nginx -s reload
```

---

## 3. Health Checks & Diagnostic Endpoints

| Component | Target / Endpoint | Expected Response | Remediation If Failing |
| :--- | :--- | :--- | :--- |
| **API Core** | `GET https://erp.example.sa/healthz` | `HTTP 200 {"status":"ok"}` | Check `erp-api` logs for DB connection failure or OOM panic. |
| **Edge Nginx** | Port 80/443 | `HTTP 200 / 301` | Verify certificate validity and Nginx syntax: `nginx -t`. |
| **PostgreSQL** | `pg_isready -U erp_user -d erp_db` | `accepting connections` | Inspect storage capacity (`df -h`) and connection count. |
| **Redis** | `redis-cli ping` | `PONG` | Check memory utilization and appendonly file corruption. |
| **Web Frontend** | `GET https://erp.example.sa/` | `HTTP 200` | Restart `erp-web` container; check Node SSR memory. |

---

## 4. Scaling Guidelines

### Horizontal Scaling
- **`erp-api`**: The Go backend is stateless. Scale to multiple instances behind Nginx or an AWS/GCP Application Load Balancer:
  ```bash
  docker compose -f docker-compose.prod.yml up -d --scale erp-api=4
  ```
- **`erp-worker`**: Workers process tasks off Redis queues. Can be scaled dynamically during high ZATCA clearance loads or payroll runs:
  ```bash
  docker compose -f docker-compose.prod.yml up -d --scale erp-worker=3
  ```

### Database Tuning (PostgreSQL 16)
Ensure `postgresql.conf` parameters match host hardware:
- `max_connections = 200`
- `shared_buffers = 25% of total RAM`
- `effective_cache_size = 75% of total RAM`
- `work_mem = 16MB`
- `maintenance_work_mem = 256MB`

---

## 5. Incident Response & Troubleshooting

### Scenario A: PostgreSQL Connection Pool Exhaustion (`too many connections`)
1. Check active connections:
   ```sql
   SELECT count(*), state FROM pg_stat_activity GROUP BY state;
   ```
2. Identify long-running transactions:
   ```sql
   SELECT pid, now() - xact_start AS duration, query 
   FROM pg_stat_activity 
   WHERE state = 'active' AND (now() - xact_start) > interval '30 seconds';
   ```
3. Terminate stuck connection:
   ```sql
   SELECT pg_terminate_backend(pid);
   ```
4. Verify `DATABASE_URL` pool size (`pool_max_conns=50`).

### Scenario B: ZATCA Phase 2 Clearance Queue Backlog
1. Inspect Redis queue length:
   ```bash
   docker compose -f docker-compose.prod.yml exec redis redis-cli -a $REDIS_PASSWORD LLEN zatca_submission_queue
   ```
2. Verify ZATCA endpoint connectivity:
   ```bash
   docker compose -f docker-compose.prod.yml exec erp-api curl -I https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation/invoices/clearance/single
   ```
3. Check worker error logs:
   ```bash
   docker compose -f docker-compose.prod.yml logs --tail=200 erp-worker | grep -i "zatca"
   ```

### Scenario C: Rate Limiting or CSRF Blocks
1. If legitimate corporate clients receive `HTTP 429 Too Many Requests`:
   - Inspect client IP against Nginx rate limiting zones (`limit_req_zone`).
   - If behind a corporate proxy or Cloudflare, ensure `real_ip_header CF-Connecting-IP;` or `X-Forwarded-For` is configured.
2. If clients receive `HTTP 403 Invalid Origin / CSRF validation failed`:
   - Check `CORS_ALLOWED_ORIGINS` in `.env.production` to ensure tenant domains are explicitly allowed.

---

## 6. On-Call Escalation Matrix

1. **Level 1 (Ops Engineer)**: Monitor alerts, execute initial runbook steps, restart crashed containers. (SLA: 15 mins)
2. **Level 2 (Lead Platform Engineer)**: Investigate DB locks, worker failures, or release rollbacks. (SLA: 30 mins)
3. **Level 3 (VP of Engineering / Security Officer)**: Security incidents, ZATCA regulatory compliance outages, critical data loss events. (SLA: Immediate)
