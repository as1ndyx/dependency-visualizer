package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// pomXML создаёт короткий POM для независимых от сети тестов этапа 2.
func pomXML(body string) string {
	return `<project xmlns="http://maven.apache.org/POM/4.0.0"><modelVersion>4.0.0</modelVersion>` + body + `</project>`
}

// dependencyXML описывает зависимость; extra содержит дополнительные поля XML.
func dependencyXML(artifact, version, extra string) string {
	return `<dependency><groupId>org.example</groupId><artifactId>` + artifact +
		`</artifactId><version>` + version + `</version>` + extra + `</dependency>`
}

// parentXML описывает родительский POM в тестовом репозитории.
func parentXML(artifact string) string {
	return `<parent><groupId>org.example</groupId><artifactId>` + artifact +
		`</artifactId><version>1</version></parent>`
}

// mavenServer публикует POM по стандартным путям и считает HTTP-запросы.
func mavenServer(t *testing.T, files map[string]string) (*httptest.Server, func() map[string]int) {
	t.Helper()
	counts := make(map[string]int)
	var mutex sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		counts[request.URL.Path]++
		mutex.Unlock()
		body, ok := files[request.URL.Path]
		if !ok {
			http.NotFound(out, request)
			return
		}
		fmt.Fprint(out, body)
	}))
	t.Cleanup(server.Close)
	return server, func() map[string]int {
		mutex.Lock()
		defer mutex.Unlock()
		result := make(map[string]int)
		for key, count := range counts {
			result[key] = count
		}
		return result
	}
}

// mavenPath возвращает путь POM для тестовых координат org.example:artifact:1.
func mavenPath(artifact string) string {
	return "/org/example/" + artifact + "/1/" + artifact + "-1.pom"
}

// writeMavenConfig создаёт INI с локальным HTTP-репозиторием.
func writeMavenConfig(t *testing.T, body string) string {
	t.Helper()
	server, _ := mavenServer(t, map[string]string{mavenPath("app"): pomXML(body),
		"/org/example/lib/2/lib-2.pom": pomXML("")})
	text := strings.Replace(validINI, "org.apache.commons:commons-lang3:3.12.0", "org.example:app:1", 1)
	text = strings.Replace(text, "https://repo.maven.apache.org/maven2", server.URL, 1)
	return writeConfig(t, text)
}
