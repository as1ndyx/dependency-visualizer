// Команда depgraph запускает инструмент визуализации зависимостей пакетов.
package main

import (
	"os"

	"dependencyvisualizer/src/app"
)

// main передаёт параметры командной строки приложению.
// Этап 1: загрузка INI-конфигурации, проверка и вывод параметров.
func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
