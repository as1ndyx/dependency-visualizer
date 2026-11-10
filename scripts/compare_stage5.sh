#!/bin/sh
# Этап 5: отдельное эталонное сравнение; Go-приложение Maven не запускает.
set -eu
cd "$(dirname "$0")/.."
for tool in curl mvn dot; do
    command -v "$tool" >/dev/null || { echo "Не найдена команда $tool" >&2; exit 1; }
done
mkdir -p bin/stage5/comparison
comparison_dir=$(cd bin/stage5/comparison && pwd)

# compare загружает исходный POM пакета и получает стандартное дерево Maven.
compare() {
    name=$1
    pom_path=$2
    project_dir="$comparison_dir/$name"
    mkdir -p "$project_dir"
    curl --fail --silent --show-error --location \
        "https://repo.maven.apache.org/maven2/$pom_path" -o "$project_dir/pom.xml"
    mvn -B -ntp -f "$project_dir/pom.xml" \
        "-Dmaven.repo.local=$comparison_dir/cache" \
        org.apache.maven.plugins:maven-dependency-plugin:3.8.1:tree \
        -DoutputType=dot "-DoutputFile=$project_dir/maven.dot"
    dot -Tsvg "$project_dir/maven.dot" -o "$project_dir/maven.svg"
    cat "$project_dir/maven.dot"
}

compare junit junit/junit/4.13.2/junit-4.13.2.pom
compare hamcrest-library org/hamcrest/hamcrest-library/1.3/hamcrest-library-1.3.pom
compare slf4j-simple org/slf4j/slf4j-simple/1.7.36/slf4j-simple-1.7.36.pom
echo "Эталонные деревья сохранены в bin/stage5/comparison/"
