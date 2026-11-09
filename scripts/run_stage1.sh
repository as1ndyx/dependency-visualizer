#!/bin/sh
# Этап 1: регрессионная проверка INI и ошибок всех параметров после добавления этапа 2.
set -eu
cd "$(dirname "$0")/.."
go build -o bin/depgraph ./src

echo "--- чтение INI в обычном и тестовом режимах ---"
go test ./tests -run '^TestLoad' -v

# expect_error проверяет сообщение и код возврата настоящего CLI.
expect_error() {
    label=$1
    shift
    echo "--- $label ---"
    status=0
    output=$(./bin/depgraph "$@" 2>&1) || status=$?
    printf '%s\n' "$output"
    if [ "$status" -ne 1 ]; then
        echo "ожидался код возврата 1, получен $status" >&2
        exit 1
    fi
    case "$output" in
        *"ошибка:"*) ;;
        *) echo "нет сообщения об ошибке" >&2; exit 1 ;;
    esac
    echo "код возврата: $status"
}

for file in examples/errors/*.ini; do
    expect_error "$file" -config "$file"
done
expect_error "конфигурация не найдена" -config examples/errors/missing-file.ini
expect_error "пустой путь к конфигурации" -config ""
expect_error "неизвестный флаг" -unknown
expect_error "лишний аргумент" extra
