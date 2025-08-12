#!/bin/bash
# dev-metrics.sh - Quick metrics check

echo "=== Prometheus Status ==="
curl -s http://localhost:9090/-/healthy

echo -e "\n=== Main Server Metrics Sample ==="
curl -s http://localhost:8080/metrics | head -20

echo -e "\n=== Generate Test Audit Events ==="
curl -X POST http://localhost:8080/audit
echo ""
