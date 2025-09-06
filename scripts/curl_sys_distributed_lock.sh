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

echo "=== SysDistributedLock API 测试 ==="
echo ""

echo "1. 尝试获取分布式锁"
curl -s -X POST "${BASE_URL}/sys_distributed_lock/acquire" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"lock_key":"test_lock","ttl":30}'
echo -e "\n"

echo "2. 查询锁信息"
curl -s -X GET "${BASE_URL}/sys_distributed_lock/query/test_lock" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "3. 检查锁是否被持有"
curl -s -X GET "${BASE_URL}/sys_distributed_lock/check/test_lock" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "4. 续期分布式锁"
curl -s -X POST "${BASE_URL}/sys_distributed_lock/renew" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"lock_key":"test_lock","ttl":60}'
echo -e "\n"

echo "5. 释放分布式锁"
curl -s -X POST "${BASE_URL}/sys_distributed_lock/release" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"lock_key":"test_lock"}'
echo -e "\n"

echo "=== SysDistributedLock API 测试完成 ==="
