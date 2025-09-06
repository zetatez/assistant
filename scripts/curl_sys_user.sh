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

echo "=== SysUser API 测试 ==="
echo ""

echo "1. 统计用户数量"
curl -s -X GET "${BASE_URL}/sys_user/count" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "2. 创建用户"
curl -s -X POST "${BASE_URL}/sys_user/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"user_name":"test_user","email":"test@example.com"}'
echo -e "\n"

echo "3. 获取用户详情"
curl -s -X GET "${BASE_URL}/sys_user/get/2" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "4. 获取用户列表"
curl -s -X GET "${BASE_URL}/sys_user/list?page=1&page_size=10" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "5. 更新用户信息"
curl -s -X PUT "${BASE_URL}/sys_user/update/3" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"user_name":"updated_user"}'
echo -e "\n"

echo "6. 删除用户 (非内置用户才能删除)"
curl -s -X DELETE "${BASE_URL}/sys_user/delete/3" -H "Authorization: Bearer ${TOKEN}"
echo -e "\n"

echo "7. 通过邮箱搜索用户"
curl -s -X POST "${BASE_URL}/sys_user/search_by_email" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"email":"admin"}'
echo -e "\n"

echo "8. 通过用户名搜索用户"
curl -s -X POST "${BASE_URL}/sys_user/search_by_user_name" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"user_name":"admin"}'
echo -e "\n"

echo "=== SysUser API 测试完成 ==="
