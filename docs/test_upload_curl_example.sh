#!/bin/bash

# Простой пример загрузки файлов через cURL
# Используйте этот скрипт или скопируйте команду

BASE_URL="http://localhost:8090"
OPERATION_KEY="op-test-$(date +%s)"

echo "Загрузка файлов:"
echo "  - docs/example.pdf"
echo "  - docs/example2.pdf"
echo ""

curl -X POST "${BASE_URL}/upload" \
  -H "X-Operation-Key: ${OPERATION_KEY}" \
  -F "files=@docs/example.pdf" \
  -F "files=@docs/example2.pdf"

echo ""
echo ""
echo "Для получения статуса используйте operation_id из ответа выше"

