# Saudi Enterprise ERP - Zero-Downtime Rolling Upgrade Guide

## 1. Upgrade Philosophy & Principles
To achieve continuous availability (99.99% uptime) in mission-critical financial, retail, and supply chain operations, upgrades must adhere to the **Expand and Contract (Parallel Run)** design pattern:
1. **Schema Backward Compatibility**: Database changes must be non-breaking to currently running older API versions.
2. **Stateless API Redundancy**: API servers are stateless and run behind an Nginx reverse proxy or ALB that performs health checks before routing traffic.
3. **Graceful Worker Draining**: Background workers complete current ZATCA clearance and payroll jobs before termination.

---

## 2. Safe Database Migration Rules
When crafting Goose migrations for production:
- **Never rename columns or tables directly**: Instead, add a new column, dual-write in application code, backfill data, and remove the old column in a subsequent release.
- **Always add new columns as nullable or with defaults**: Avoid long table locks on multi-million row tables.
- **Create indexes concurrently**: Use `CREATE INDEX CONCURRENTLY` in PostgreSQL to prevent read/write locks on transaction tables.
- **Pre-deployment Migration**: Always apply non-breaking migrations *before* switching application traffic to the new image.

---

## 3. Rolling Upgrade Step-by-Step Procedure

### Phase 1: Pre-Upgrade Verification & Backup
```bash
# 1. Execute automated safety backup
./scripts/backup.sh

# 2. Run system doctor to ensure healthy baseline
./scripts/doctor.sh

# 3. Pull new container images
docker compose -f deployments/compose/docker-compose.prod.yml pull
```

### Phase 2: Database Schema Upgrade (Expand Phase)
```bash
# Apply additive database migrations
docker compose -f deployments/compose/docker-compose.prod.yml run --rm erp-migrate goose -dir /database/migrations postgres "$DATABASE_URL" up
```

### Phase 3: Rolling Application Deployment
```bash
# 1. Update and restart API instances one by one (or rolling update)
docker compose -f deployments/compose/docker-compose.prod.yml up -d --no-deps erp-api

# 2. Wait for API healthcheck to turn healthy
until [ "$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/healthz)" == "200" ]; do
    echo "Waiting for erp-api to become healthy..."
    sleep 2
done

# 3. Update Web frontend
docker compose -f deployments/compose/docker-compose.prod.yml up -d --no-deps erp-web

# 4. Update Background workers
docker compose -f deployments/compose/docker-compose.prod.yml up -d --no-deps erp-worker

# 5. Gracefully reload Nginx routing tables
docker compose -f deployments/compose/docker-compose.prod.yml exec nginx nginx -s reload
```

### Phase 4: Post-Deployment Smoke Testing
Verify key functional pathways:
1. `GET /healthz` returns `HTTP 200`.
2. Login endpoint authentication succeeds and issues valid session tokens.
3. Query dashboard metrics endpoint.
4. Check error rates in Prometheus/Grafana.

---

## 4. Emergency Rollback Plan
If critical issues or regressions are discovered within the 15-minute post-deployment window:
1. **Revert Application Containers**:
   ```bash
   # Re-tag and deploy previous stable image tag
   PREV_TAG="v0.9.9"
   docker compose -f deployments/compose/docker-compose.prod.yml up -d --no-deps erp-api erp-web erp-worker
   ```
2. **Database Rollback (if applicable)**:
   - If migration was purely additive (recommended), no DB rollback is needed; older code will simply ignore new columns.
   - If emergency DB rollback is mandatory:
     ```bash
     docker compose -f deployments/compose/docker-compose.prod.yml run --rm erp-migrate goose -dir /database/migrations postgres "$DATABASE_URL" down
     ```
3. **Notify Incident Command**: Post update in `#incident-management` and document root cause.
