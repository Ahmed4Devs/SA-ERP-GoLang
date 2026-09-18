# Production Monitoring & Observability Guide

## 1. Overview
The Saudi Enterprise ERP observability stack is built around Prometheus metrics collection, Grafana dashboard visualization, and Alertmanager incident routing.

The API exports Prometheus metrics. PostgreSQL, Nginx, container, and worker telemetry require the separately deployed exporters/collector described below; they are not bundled in the production Compose file.

---

## 2. Core Metrics Inventory

### Application & API Metrics (Go Backend)
| Metric Name | Type | Description |
| :--- | :--- | :--- |
| `http_requests_total{method, path, status}` | Counter | Total HTTP requests categorized by HTTP status code |
| `http_request_duration_seconds{method, path}` | Histogram | Request latency distributions (p50, p95, p99) |
| `db_pool_connections_active` | Gauge | Active connections currently leased from database pool |
| `db_pool_connections_idle` | Gauge | Idle connections available in the pool |
| `db_pool_connections_wait_duration_seconds` | Counter | Duration queries waited for an available connection |

### Regulatory & Business Compliance Metrics (Saudi ERP)
| Metric Name | Type | Description |
| :--- | :--- | :--- |
| `zatca_invoices_cleared_total` | Counter | Total standard tax invoices successfully cleared with ZATCA |
| `zatca_invoices_reported_total` | Counter | Total simplified tax invoices reported with cryptographic stamp |
| `zatca_submissions_failed_total{error_code}` | Counter | Invoices rejected by ZATCA compliance API |
| `wps_payroll_batches_processed_total` | Counter | Total Wages Protection System (Mudad) files processed |
| `gosi_monthly_contributions_calculated_total` | Counter | Total employee GOSI social insurance deduction jobs |

### Infrastructure & System Metrics
| Component | Metric Source | Key Telemetry |
| :--- | :--- | :--- |
| **Nginx** | `nginx-prometheus-exporter` | Active connections, request rate, upstream latency |
| **PostgreSQL** | `postgres_exporter` | Active transactions, cache hit ratio, deadlocks, disk usage |
| **Containers** | `cAdvisor` | CPU % utilization, memory RSS, container restart counts |

---

## 3. Recommended Alert Rules (Prometheus Alertmanager)

```yaml
groups:
  - name: erp_critical_alerts
    rules:
      # 1. High API 5xx Error Rate
      - alert: HighErrorRate
        expr: (sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))) * 100 > 2
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "API 5xx error rate exceeded 2% (current: {{ $value }}%)"
          runbook_url: "https://docs.erp.example.sa/operations/runbook#high-error-rate"

      # 2. Database Connection Exhaustion
      - alert: DatabasePoolNearlyFull
        expr: (db_pool_connections_active / 50) * 100 > 85
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "PostgreSQL connection pool utilization > 85%"

      # 3. ZATCA Compliance Failures
      - alert: HighZatcaInvoiceRejectionRate
        expr: sum(rate(zatca_submissions_failed_total[10m])) > 5
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "ZATCA clearance/reporting rejections spiked in the last 10 minutes"
          runbook_url: "https://docs.erp.example.sa/operations/runbook#zatca-backlog"

      # 4. API Latency Degraded (p95 > 1s)
      - alert: HighApiLatencyP95
        expr: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le)) > 1.0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "API p95 latency exceeded 1.0s"

      # 5. Disk Space Low (PostgreSQL Data Volume)
      - alert: DiskSpaceLow
        expr: (node_filesystem_free_bytes{mountpoint="/var/lib/postgresql/data"} / node_filesystem_size_bytes) * 100 < 15
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "PostgreSQL volume free space is below 15%"
```

---

## 4. Prometheus Scrape Configuration (`prometheus.yml`)

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'erp-api'
    metrics_path: '/metrics'
    static_configs:
      - targets: ['erp-api:8080']

  - job_name: 'erp-postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'erp-nginx'
    static_configs:
      - targets: ['nginx-exporter:9113']
```

---

## 5. Grafana Dashboard Panels
Recommended default panels for the Executive SRE Dashboard:
1. **System Health & RPS**: Total requests/sec, 2xx/4xx/5xx breakdown.
2. **P95 / P99 Latency**: Microservice and endpoint breakdown.
3. **ZATCA Real-Time Throughput**: Standard vs. simplified invoices per minute.
4. **Active Tenant Count & Legal Entities**: Tenant distribution and concurrency.
5. **Database Query Performance**: Slow query logs (>200ms) and lock duration.
