// Команда depgraph запускает инструмент визуализации зависимостей пакетов.
package main

import (
	"os"

	"dependencyvisualizer/src/app"
)

// main передаёт параметры командной строки приложению.
// Этап 3: загрузка конфигурации и построение полного графа рекурсивным DFS.
func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
