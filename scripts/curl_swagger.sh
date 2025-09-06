#!/bin/bash

# Swagger API 访问脚本
BASE_URL="http://127.0.0.1:9876"

echo "=== Swagger API 文档 ==="
echo ""

echo "Swagger API 文档界面:"
echo "curl -X GET ${BASE_URL}/swagger/index.html"
curl -X GET "${BASE_URL}/swagger/index.html"
echo -e "\n"

echo "Swagger JSON (OpenAPI规范):"
echo "curl -X GET ${BASE_URL}/swagger/doc.json"
curl -X GET "${BASE_URL}/swagger/doc.json"
echo -e "\n"

echo "Swagger YAML (OpenAPI规范):"
echo "curl -X GET ${BASE_URL}/swagger/doc.yaml"
curl -X GET "${BASE_URL}/swagger/doc.yaml"
echo -e "\n"

echo "=== Swagger API 文档访问完成 ==="
