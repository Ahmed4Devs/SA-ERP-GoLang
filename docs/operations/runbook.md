# Saudi Enterprise ERP - Production Operations Runbook

## 1. Overview & Architecture Topology
This document provides standard operating procedures (SOP) for the 24/7 operations, maintenance, and incident response of the Saudi Enterprise ERP production system.

### Container Topology
The production deployment starts five long-running services, three setup jobs,
and an optional backup job via Docker Compose:
- **`nginx`**: Edge reverse proxy handling TLS termination, rate limiting, and HTTP/2 routing.
- **`erp-web`**: Next.js 16 standalone frontend web server (Port 3000).
- **`erp-api`**: Go core backend API service with pgx connection pooling (Port 8080).
- **`erp-worker`**: Go background worker that dispatches the transactional outbox, delivers queued email notifications, and—when enabled—retries pending ZATCA submissions.
- **`erp-migrate`**: Ephemeral schema migration runner executing Goose database migrations before API launch.
- **`erp-provision`** and **`erp-grants`**: Provision database roles and grant runtime privileges before API launch. See [database identities](database-roles.md).
- **`postgres`**: PostgreSQL 18 relational database with Saudi localized schema (`Asia/Riyadh` TZ).
- **`erp-backup`**: Encrypted database + attachment recovery job in the `operations` profile; invoke through `scripts/backup-coordinated.sh`. A systemd timer template is included.

---

## 2. Service Lifecycle Commands

### Starting the Production Stack
```bash
# Navigate to deployment directory
cd deployments/compose

# Validate environment configuration
cp ../../.env.production.example .env.production
# Ensure secrets and credentials are fully set

# Pull & build containers with zero cache
docker compose --env-file .env.production -f docker-compose.prod.yml pull
docker compose --env-file .env.production -f docker-compose.prod.yml build --no-cache

# Start stack in background
docker compose --env-file .env.production -f docker-compose.prod.yml up -d
```

For first installation, set `BOOTSTRAP_ADMIN_ENABLED=true`, provide the operator's
email and a unique password of at least 16 characters, and leave demo data disabled.
After the first successful login, set `BOOTSTRAP_ADMIN_ENABLED=false`, remove
`BOOTSTRAP_ADMIN_PASSWORD` from the environment file, and recreate `erp-api`.
Bootstrap does not reset an existing user's password.

Set `APP_URL` to the actual HTTPS origin, update the Nginx `server_name` in
`deployments/nginx/conf.d/erp.conf`, and install the matching certificate chain
and private key at `TLS_CERT_PATH` / `TLS_KEY_PATH`. Restrict host-file access to
the operator and container runtime. Configure certificate renewal externally;
after renewal run `docker compose ... exec nginx nginx -t` followed by a reload.
The Compose stack does not issue certificates automatically.

### Checking Stack Status & Health
```bash
# View running status and container health
docker compose --env-file .env.production -f docker-compose.prod.yml ps

# Follow logs across all services
docker compose --env-file .env.production -f docker-compose.prod.yml logs -f --tail=100

# Inspect specific service logs (e.g. API or Worker)
docker compose --env-file .env.production -f docker-compose.prod.yml logs -f erp-api
docker compose --env-file .env.production -f docker-compose.prod.yml logs -f erp-worker
```

### Stopping & Graceful Restart
```bash
# Graceful restart of API service
docker compose --env-file .env.production -f docker-compose.prod.yml restart erp-api

# Stop full stack without data deletion
docker compose --env-file .env.production -f docker-compose.prod.yml down

# Graceful reload of Nginx configurations without dropping connections
docker compose --env-file .env.production -f docker-compose.prod.yml exec nginx nginx -s reload
```

---

## 3. Health Checks & Diagnostic Endpoints

| Component | Target / Endpoint | Expected Response | Remediation If Failing |
| :--- | :--- | :--- | :--- |
| **API Core** | `GET https://erp.example.sa/healthz` | `HTTP 200 {"status":"ok"}` | Check `erp-api` logs for DB connection failure or OOM panic. |
| **API Readiness** | `GET https://erp.example.sa/readyz` | `HTTP 200` | Check migration version, database connectivity, and document storage. |
| **Edge Nginx** | Port 80/443 | `HTTP 200 / 301` | Verify certificate validity and Nginx syntax: `nginx -t`. |
| **PostgreSQL** | `pg_isready -U erp_user -d erp_db` | `accepting connections` | Inspect storage capacity (`df -h`) and connection count. |
| **Web Frontend** | `GET https://erp.example.sa/` | `HTTP 200` | Restart `erp-web` container; check Node SSR memory. |

---

## 4. Scaling Guidelines

### Horizontal Scaling
- **`erp-api`**: API request processing is stateless, but uploaded documents use the Compose `documentdata` volume. Multiple instances on one Docker host can share that volume. Cross-host scaling requires a reviewed shared filesystem or object-storage implementation before rollout:
  ```bash
  docker compose --env-file .env.production -f docker-compose.prod.yml up -d --scale erp-api=4
  ```
- **`erp-worker`**: Workers lease outbox rows with `FOR UPDATE SKIP LOCKED`, so notification delivery can scale horizontally. SMTP is an at-least-once external transport; recipients can see a duplicate only in the narrow case where SMTP accepts a message and the worker crashes before recording success:
  ```bash
  docker compose --env-file .env.production -f docker-compose.prod.yml up -d --scale erp-worker=3
  ```

### Database Tuning (PostgreSQL 18)
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

### Scenario B: ZATCA Submission Retry Backlog
1. Inspect failed and pending submissions in PostgreSQL and correlate them with worker logs.
2. Verify ZATCA endpoint connectivity from an approved diagnostic environment:
   ```bash
   docker compose --env-file .env.production -f docker-compose.prod.yml exec erp-api curl -I https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation/invoices/clearance/single
   ```
3. Check worker error logs:
   ```bash
   docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=200 erp-worker | grep -i "zatca"
   ```

### Scenario C: Notification or Outbox Backlog
1. Inspect ready, leased, and dead-letter events:
   ```sql
   SELECT event_type, attempt_count, available_at, dead_lettered_at, last_error
   FROM platform.outbox_events
   WHERE published_at IS NULL
   ORDER BY occurred_at;
   ```
2. Verify the worker is running even when `ZATCA_ENABLED=false`, then inspect dispatch errors:
   ```bash
   docker compose --env-file .env.production -f docker-compose.prod.yml ps erp-worker
   docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=200 erp-worker | grep -i "outbox\|notification"
   ```
3. Resolve SMTP credentials/connectivity before re-queuing a dead-letter row. Re-queue only the exact reviewed event:
   ```sql
   UPDATE platform.outbox_events
   SET dead_lettered_at = NULL, attempt_count = 0, last_error = NULL, available_at = now()
   WHERE id = '<reviewed-event-uuid>' AND published_at IS NULL;
   ```

### Scenario D: Rate Limiting or CSRF Blocks
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

## Recovery and regulatory operation

See [coordinated backups](../backup/backup-strategy.md), [restore procedure](../restore/restore-guide.md), and [regulatory integration boundaries](regulatory-integrations.md) before deployment.
