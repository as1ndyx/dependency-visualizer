#!/bin/sh
# Этап 5: три реальных пакета, текст Graphviz на экране и диаграммы SVG.
set -eu
cd "$(dirname "$0")/.."
command -v dot >/dev/null || { echo "Нужен Graphviz: команда dot" >&2; exit 1; }
go build -o bin/depgraph ./src
mkdir -p bin/stage5
for name in junit hamcrest-library slf4j-simple; do
    echo "--- $name ---"
    ./bin/depgraph -config "examples/stage5/$name.ini" > "bin/stage5/$name.dot"
    cat "bin/stage5/$name.dot"
    dot -Tsvg "bin/stage5/$name.dot" -o "bin/stage5/$name.svg"
done
echo "Диаграммы сохранены в bin/stage5/"
