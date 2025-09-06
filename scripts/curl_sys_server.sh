#!/bin/bash

BASE_URL="http://127.0.0.1:9876"

echo "=== 登录获取 Token ==="
LOGIN_RESP=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"AAaa00__"}')

TOKEN=$(echo $LOGIN_RESP | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "登录失败: $LOGIN_RESP"
  exit 1
fi

echo "Token: $TOKEN"
echo ""

echo "=== SysServer API 测试 ==="
echo ""

echo "1. 统计服务器数量"
curl -s -X GET "${BASE_URL}/sys_server/count" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "2. 获取服务器详情"
curl -s -X GET "${BASE_URL}/sys_server/get/1" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "3. 获取服务器列表"
curl -s -X GET "${BASE_URL}/sys_server/list?page=1&page_size=10" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "4. 通过IDC搜索服务器"
curl -s -X POST "${BASE_URL}/sys_server/search_by_idc" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"idc":"aws"}'
echo -e "\n"

echo "5. 通过IP地址搜索服务器"
curl -s -X POST "${BASE_URL}/sys_server/search_by_svr_ip" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"svr_ip":"192.168.1.1"}'
echo -e "\n"

echo "6. 通过IDC和IP组合搜索"
curl -s -X POST "${BASE_URL}/sys_server/search_by_idc_and_svr_ip" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"idc":"aws","svr_ip":"192.168.1.1"}'
echo -e "\n"

echo "=== SysServer API 测试完成 ==="
