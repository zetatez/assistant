#!/bin/bash

BASE_URL="http://127.0.0.1:9876"

echo "=== Health API 测试 ==="
echo ""

echo "1. 服务健康检查 (无需认证)"
curl -s "${BASE_URL}/health" | python3 -m json.tool 2>/dev/null || curl -s "${BASE_URL}/health"
echo ""

echo "=== Health API 测试完成 ==="
