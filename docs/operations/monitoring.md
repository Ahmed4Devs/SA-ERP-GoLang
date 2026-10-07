# Production monitoring

The API exposes Prometheus metrics at `/metrics` on its internal port 8080.
Production requires `Authorization: Bearer <METRICS_BEARER_TOKEN>`. Nginx does
not expose this endpoint publicly. Prometheus, Grafana, Alertmanager, database
exporters, and host exporters must be installed separately; they are not part of
the checked-in Compose deployment.

## Metrics currently emitted

| Metric | Type | Labels |
| --- | --- | --- |
| `erp_http_requests_total` | Counter | `method`, `path`, `status` |
| `erp_http_request_duration_seconds` | Histogram | `method`, `path`, `status` |
| `erp_http_requests_in_flight` | Gauge | None |

Paths use route patterns to avoid a separate series for every record identifier.
Database-pool, ZATCA, payroll, and GOSI business counters are not currently
emitted by the API. Do not configure alerts against nonexistent series.

## Scraping

Connect the collector to the deployment's private network. Mount a secret file
containing the same value as `METRICS_BEARER_TOKEN` into the collector only:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
scrape_configs:
  - job_name: erp-api
    metrics_path: /metrics
    authorization:
      type: Bearer
      credentials_file: /run/secrets/erp_metrics_token
    static_configs:
      - targets: ['erp-api:8080']
```

## Alerts supported by current metrics

```yaml
groups:
  - name: erp_api
    rules:
      - alert: ERPAPIUnavailable
        expr: up{job="erp-api"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: ERP API metrics cannot be scraped
      - alert: ERPAPIHighErrorRate
        expr: sum(rate(erp_http_requests_total{status=~"5.."}[5m])) / clamp_min(sum(rate(erp_http_requests_total[5m])), 0.001) > 0.02
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: ERP API errors exceed two percent
      - alert: ERPAPIHighLatency
        expr: histogram_quantile(0.95, sum(rate(erp_http_request_duration_seconds_bucket[5m])) by (le)) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: ERP API p95 latency exceeds one second
```

Also monitor `/readyz` with an HTTP probe: successful metric scraping alone does
not prove the database schema and document storage are ready. Add host disk,
PostgreSQL connection/lock, backup age, certificate expiry, process restart,
outbox pending/dead-letter, and ZATCA retry-backlog checks using the selected
exporters or scheduled SQL checks. The runbook provides backlog queries.

Dashboards should show availability, request rate, error rate, latency, and
concurrent requests. Add business-throughput panels only after their actual
instrumentation has been implemented and verified. Test alert delivery through
the deployment's chosen incident channel before go-live.
