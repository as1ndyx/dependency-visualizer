// Команда depgraph запускает инструмент визуализации зависимостей пакетов.
package main

import (
	"os"

	"dependencyvisualizer/src/app"
)

// main передаёт параметры командной строки приложению.
// Этап 4: построение графа и режим вывода обратных зависимостей.
func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
