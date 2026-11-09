#!/bin/sh
# Этап 2: прямые зависимости трёх пакетов и ошибка получения POM.
# Для реальных запросов требуется доступ к Maven Central.
set -eu
cd "$(dirname "$0")/.."
go build -o bin/depgraph ./src

for file in examples/stage2/*.ini; do
    echo "--- $file ---"
    ./bin/depgraph -config "$file"
done

echo "--- отсутствующий пакет: HTTP 404 ---"
status=0
output=$(./bin/depgraph -config examples/missing-package.ini 2>&1) || status=$?
printf '%s\n' "$output"
if [ "$status" -ne 1 ]; then
    echo "ожидался код возврата 1, получен $status" >&2
    exit 1
fi
case "$output" in
    *"HTTP 404"*) echo "код возврата: $status" ;;
    *) echo "ожидалась ошибка HTTP 404" >&2; exit 1 ;;
esac
