#!/bin/sh
# Этап 3: рекурсивный DFS, общие зависимости, циклы, фильтрация и ошибки тестового файла.
# Все примеры работают без интернета и без установленного Maven.
set -eu
cd "$(dirname "$0")/.."
go build -o bin/depgraph ./src

for file in examples/stage3/*.ini; do
    echo "--- $file ---"
    ./bin/depgraph -config "$file"
done

for file in examples/stage3/errors/*.ini; do
    echo "--- ошибка: $file ---"
    status=0
    output=$(./bin/depgraph -config "$file" 2>&1) || status=$?
    printf '%s\n' "$output"
    if [ "$status" -ne 1 ]; then
        echo "ожидался код возврата 1, получен $status" >&2
        exit 1
    fi
    case "$output" in
        *"ошибка:"*) echo "код возврата: $status" ;;
        *) echo "нет сообщения об ошибке" >&2; exit 1 ;;
    esac
done
