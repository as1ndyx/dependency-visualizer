package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dependencyvisualizer/src/maven"
)

func TestMavenInvalidPOM(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"xml", "<project>", "XML"}, {"html", "<html/>", "XML"},
		{"two-roots", "<project/><project/>", "лишние данные"},
		{"tail", "<project/>garbage", "лишние данные"},
		{"tail-xml", "<project/><", "конец XML"},
		{"version", pomXML("<dependencies>" + dependencyXML("lib", "", "") + "</dependencies>"), "версию"},
		{"name", pomXML("<dependencies>" + dependencyXML("", "1", "") + "</dependencies>"), "имя зависимости"},
		{"scope", pomXML("<dependencies>" + dependencyXML("lib", "1", "<scope>x</scope>") +
			"</dependencies>"), "scope"},
		{"optional", pomXML("<dependencies>" + dependencyXML("lib", "1", "<optional>x</optional>") +
			"</dependencies>"), "optional"},
		{"unknown-property", pomXML("<dependencies>" + dependencyXML("lib", "${missing}", "") +
			"</dependencies>"), "неизвестное свойство"},
		{"incomplete-property", pomXML("<dependencies>" + dependencyXML("lib", "${missing", "") +
			"</dependencies>"), "незавершённая ссылка"},
		{"property-cycle", pomXML("<properties><x>${x}${x}</x></properties><dependencies>" +
			dependencyXML("lib", "${x}", "") + "</dependencies>"), "цикл"},
		{"bom-type", pomXML("<dependencyManagement><dependencies>" +
			dependencyXML("bom", "1", "<scope>import</scope>") + "</dependencies></dependencyManagement>"), "type=pom"},
		{"large", strings.Repeat("x", maven.MaxPOMBytes+1), "размер"},
		{"charset", `<?xml version="1.0" encoding="unknown"?><project/>`, "кодировка"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, _ := mavenServer(t, map[string]string{mavenPath("app"): tc.body})
			_, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ожидали %q, получили %v", tc.want, err)
			}
		})
	}
}

func TestMavenCycles(t *testing.T) {
	for _, kind := range []string{"parent", "bom"} {
		t.Run(kind, func(t *testing.T) {
			body := parentXML("app")
			if kind == "bom" {
				body = "<dependencyManagement><dependencies>" +
					dependencyXML("app", "1", "<scope>import</scope><type>pom</type>") +
					"</dependencies></dependencyManagement>"
			}
			server, _ := mavenServer(t, map[string]string{mavenPath("app"): pomXML(body)})
			_, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
			if err == nil || !strings.Contains(err.Error(), "цикл") {
				t.Fatalf("ожидали ошибку цикла %s, получили %v", kind, err)
			}
		})
	}
}

func TestMavenHTTPStatus(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, request *http.Request) {
				out.WriteHeader(status)
			}))
			defer server.Close()
			_, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
			if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", status)) {
				t.Fatalf("ожидали HTTP %d, получили %v", status, err)
			}
		})
	}
}

func TestMavenTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()
	client := maven.NewClient(&http.Client{Timeout: 30 * time.Millisecond})
	_, err := client.DirectDependencies(context.Background(), server.URL, "org.example:app:1")
	if err == nil || !strings.Contains(err.Error(), "не удалось получить POM") {
		t.Fatalf("ожидали ошибку тайм-аута, получили %v", err)
	}
}

func TestMavenReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, request *http.Request) {
		out.Header().Set("Content-Length", "100")
		fmt.Fprint(out, "<project/>")
	}))
	defer server.Close()
	_, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
	if err == nil || !strings.Contains(err.Error(), "ошибка чтения") {
		t.Fatalf("ожидали ошибку неполного ответа, получили %v", err)
	}
}

func TestMavenModelLimit(t *testing.T) {
	files := make(map[string]string)
	for i := 0; i <= maven.MaxModels; i++ {
		name := fmt.Sprintf("parent%d", i)
		files[mavenPath(name)] = pomXML(parentXML(fmt.Sprintf("parent%d", i+1)))
	}
	server, _ := mavenServer(t, files)
	_, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:parent0:1")
	if err == nil || !strings.Contains(err.Error(), "предел") {
		t.Fatalf("ожидали ограничение числа POM, получили %v", err)
	}
}

func TestMavenInvalidAddress(t *testing.T) {
	coordinates := maven.Coordinates{GroupID: "org.example", ArtifactID: "lib", Version: "1"}
	for _, address := range []string{"", "://bad", "file:///tmp", "https://host?x=1", "https://user:pass@host"} {
		if _, err := maven.POMURL(address, coordinates); err == nil {
			t.Errorf("принят неверный адрес %q", address)
		}
	}
	for _, name := range []string{"", "group:artifact", "org..example:lib:1", "g:../lib:1", "g:lib:1-SNAPSHOT"} {
		if _, err := maven.NewClient(nil).DirectDependencies(context.Background(), "https://host", name); err == nil {
			t.Errorf("приняты неверные координаты %q", name)
		}
	}
}
