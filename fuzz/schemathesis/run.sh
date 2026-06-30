#!/usr/bin/env bash
# fuzz/schemathesis/run.sh
#
# Прогон чёрно-ящичного fuzz-тестирования через Schemathesis.
# Воспроизводит результаты, описанные в отчёте: 16 эндпоинтов × 100 запросов = 1600.
#
# Schemathesis генерирует запросы из OpenAPI-схемы по property-based-стратегиям
# (hypothesis): валидные/невалидные значения, граничные числа, не-UTF-8 байты,
# глубокая вложенность, пустые объекты и т.п.
#
# ТРЕБОВАНИЯ:
#   - запущенный docker-compose с поднятыми всеми сервисами и gateway на :8080;
#   - swagger-схемы каждого сервиса доступны по /swagger/doc.json;
#   - Python 3.10+ и schemathesis: pip install schemathesis==3.36.* (или новее).
#
# USAGE:
#   ./fuzz/schemathesis/run.sh           # 100 req per endpoint, как в отчёте
#   N=200 ./fuzz/schemathesis/run.sh     # переопределить число запросов
#
# Артефакты — в fuzz/reports/schemathesis-<timestamp>/.

set -euo pipefail

N="${N:-100}"
GATEWAY="${GATEWAY:-http://localhost:8080}"
AUTH_SCHEMA="${AUTH_SCHEMA:-http://localhost:8081/swagger/doc.json}"
GROUPS_SCHEMA="${GROUPS_SCHEMA:-http://localhost:8082/swagger/doc.json}"
PROJECTS_SCHEMA="${PROJECTS_SCHEMA:-http://localhost:8083/swagger/doc.json}"

cd "$(dirname "$0")/../.."
TS="$(date +%Y%m%d-%H%M%S)"
OUT="fuzz/reports/schemathesis-$TS"
mkdir -p "$OUT"

echo "→ artifacts dir: $OUT"
echo "→ N=$N requests per endpoint"
echo "→ gateway: $GATEWAY"

# Получаем валидный JWT admin'а — нужен schemathesis-у, чтобы дёргать
# защищённые эндпоинты. Регистрируемся в auth-service напрямую (если уже
# зарегистрирован — логинимся; ошибка register идёт в /dev/null).
TOKEN_FILE="$OUT/token.json"
{
  curl -fsS -X POST "$GATEWAY/auth/register" \
       -H 'Content-Type: application/json' \
       -d '{"email":"fuzz@example.com","password":"fuzz-Password-1","full_name":"Fuzz Admin","role":"admin"}' \
       2>/dev/null \
    || curl -fsS -X POST "$GATEWAY/auth/login" \
         -H 'Content-Type: application/json' \
         -d '{"email":"fuzz@example.com","password":"fuzz-Password-1"}';
} > "$TOKEN_FILE"

TOKEN="$(python3 -c "import json,sys; print(json.load(open('$TOKEN_FILE'))['access_token'])")"
echo "→ got JWT (sha256 = $(printf '%s' "$TOKEN" | sha256sum | cut -c1-12))"

# Общие флаги. --checks all даёт все встроенные проверки (server_error,
# status_code_conformance, response_schema_conformance, ...).
# --hypothesis-max-examples — лимит запросов на КАЖДЫЙ эндпоинт.
COMMON_FLAGS=(
  --checks all
  --hypothesis-max-examples "$N"
  --header "Authorization: Bearer $TOKEN"
  --base-url "$GATEWAY"
  --request-timeout 5
  --workers 4
  --cassette-path "$OUT/cassette.yaml"   # перепишется на каждом run_service
)

run_service() {
  local name="$1" schema="$2"
  echo
  echo "============================================================"
  echo "→ schemathesis run: $name"
  echo "============================================================"
  local junit="$OUT/$name.junit.xml"
  local cassette="$OUT/$name.cassette.yaml"
  local log="$OUT/$name.log"

  schemathesis run \
    "$schema" \
    --checks all \
    --hypothesis-max-examples "$N" \
    --header "Authorization: Bearer $TOKEN" \
    --base-url "$GATEWAY" \
    --request-timeout 5 \
    --workers 4 \
    --junit-xml "$junit" \
    --cassette-path "$cassette" \
    2>&1 | tee "$log" || true
  # Не падаем по exit-code: даже если найдена аномалия, нам нужны артефакты.
}

run_service auth "$AUTH_SCHEMA"
run_service groups "$GROUPS_SCHEMA"
run_service projects "$PROJECTS_SCHEMA"

# Сводный отчёт.
python3 fuzz/schemathesis/summarize.py "$OUT" > "$OUT/summary.txt"
cat "$OUT/summary.txt"

echo
echo "→ done. Artifacts: $OUT"
