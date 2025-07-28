#!/bin/bash

echo "Starting Funding Monitor..."
echo "Web interface will be available at: http://localhost:8080"
echo "API endpoints:"
echo "  - GET /api/funding (all funding rates)"
echo "  - GET /api/funding/{exchange} (exchange-specific rates)"
echo "  - GET /api/health (health check)"
echo ""
echo "Press Ctrl+C to stop the application"
echo ""

./fundingmonitor 