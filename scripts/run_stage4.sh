#!/bin/sh
# Этап 4: прямые и транзитивные обратные зависимости с использованием DFS этапа 3.
# Все случаи проверяются на тестовых файлах без интернета и Maven.
set -eu
cd "$(dirname "$0")/.."
go build -o bin/depgraph ./src

# show демонстрирует один вариант графа и выбранный пакет.
show() {
    echo "--- $1, обратные зависимости $2 ---"
    ./bin/depgraph -config "examples/stage4/$1.ini" -reverse "$2"
}

show chain D
show diamond D
show other-branch D
show cycle A
show self-cycle A
show no-dependents A
show filter D
show alternate-path D

# expect_error проверяет код возврата и пояснение для неверного запроса.
expect_error() {
    label=$1
    expected=$2
    shift 2
    echo "--- $label ---"
    task_exit_code=0
    output=$(./bin/depgraph "$@" 2>&1) || task_exit_code=$?
    printf '%s\n' "$output"
    if [ "$task_exit_code" -ne 1 ]; then
        echo "ожидался код 1, получен $task_exit_code" >&2
        exit 1
    fi
    case "$output" in
        *"$expected"*) echo "код возврата: $task_exit_code" ;;
        *) echo "не найдено ожидаемое сообщение: $expected" >&2; exit 1 ;;
    esac
}

expect_error "неизвестный пакет" "не найден в области поиска" \
    -config examples/stage4/chain.ini -reverse MISSING
expect_error "исключённый пакет" "исключён фильтром" \
    -config examples/stage4/filtered-target.ini -reverse D
expect_error "пустой пакет" "имя пакета не должно быть пустым" \
    -config examples/stage4/chain.ini -reverse ""
