#!/bin/sh
# Этап 3: реальный граф Maven и применение фильтра до загрузки зависимостей.
# Для выполнения требуется доступ к Maven Central.
set -eu
cd "$(dirname "$0")/.."
go build -o bin/depgraph ./src

echo "--- полный граф JUnit ---"
./bin/depgraph -config config.ini
echo "--- исключение hamcrest-library ---"
./bin/depgraph -config examples/maven-filter.ini
