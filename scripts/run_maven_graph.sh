#!/bin/sh
# Этап 3: реальный граф Maven и применение фильтра до загрузки зависимостей.
# Для выполнения требуется доступ к Maven Central.
set -eu
cd "$(dirname "$0")/.."
go build -o bin/depgraph ./src

echo "--- полный граф JUnit ---"
./bin/depgraph -format text -config config.ini
echo "--- исключение hamcrest-library ---"
./bin/depgraph -format text -config examples/maven-filter.ini
