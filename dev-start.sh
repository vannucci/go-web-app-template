#!/bin/bash
# dev-start.sh - Start all supporting services

echo "Starting supporting services..."
docker-compose up -d onboarding-server static-server postgres prometheus grafana

echo "Waiting for services to be ready..."
sleep 15

echo "Services started! You can now work on your main server."
echo ""
echo "Available services:"
echo "- Main server: docker-compose up main-server"
echo "- Prometheus: http://localhost:9090"
echo "- Grafana: http://localhost:3000 (admin/admin123)"
echo "- Static files: http://localhost:8082"
echo "- Onboarding API: http://localhost:8081"
echo ""
echo "To stop all services: docker-compose down"
