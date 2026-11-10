package tests

import (
	"context"
	"reflect"
	"testing"

	"dependencyvisualizer/src/maven"
)

func TestMavenDependencyInheritance(t *testing.T) {
	const expectedDependencyCount = 2 // Обычный артефакт и test-jar с classifier tests.
	parent := `<dependencies>` + dependencyXML("lib", "2", "<scope>test</scope>") + `</dependencies>`
	root := parentXML("parent") + `<dependencies>` + dependencyXML("lib", "3", "") +
		dependencyXML("lib", "4", "<type>test-jar</type><classifier>tests</classifier>") + `</dependencies>`
	server, _ := mavenServer(t, map[string]string{mavenPath("app"): pomXML(root), mavenPath("parent"): pomXML(parent)})
	got, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != expectedDependencyCount ||
		got[0].Version != "3" || got[0].Scope != "test" || got[1].Classifier != "tests" {
		t.Fatalf("неверное наследование или потерян classifier: %+v", got)
	}
}

func TestMavenBOMPrecedence(t *testing.T) {
	importBOM := func(name string) string { return dependencyXML(name, "1", "<type>pom</type><scope>import</scope>") }
	root := `<dependencyManagement><dependencies>` + importBOM("first") + importBOM("second") +
		importBOM("first") + `</dependencies></dependencyManagement><dependencies>` +
		dependencyXML("lib", "", "") + `</dependencies>`
	management := func(version string) string {
		return pomXML(`<dependencyManagement><dependencies>` + dependencyXML("lib", version, "") +
			`</dependencies></dependencyManagement>`)
	}
	server, requests := mavenServer(t, map[string]string{
		mavenPath("app"): pomXML(root), mavenPath("first"): management("2"), mavenPath("second"): management("3"),
	})
	got, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
	if err != nil || len(got) != 1 || got[0].Version != "2" {
		t.Fatalf("неверный приоритет BOM: %+v, %v", got, err)
	}
	want := map[string]int{mavenPath("app"): 1, mavenPath("first"): 1, mavenPath("second"): 1}
	if !reflect.DeepEqual(requests(), want) {
		t.Fatalf("BOM загружается повторно: %v", requests())
	}
}

func TestMavenLatin1AndXMLTail(t *testing.T) {
	body := `<?xml version="1.0" encoding="ISO-8859-1"?>` +
		pomXML("<description>caf\xe9</description>") + "\n<!-- comment --><?ok value?>\n"
	server, _ := mavenServer(t, map[string]string{mavenPath("app"): body})
	got, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
	if err != nil || len(got) != 0 {
		t.Fatalf("POM ISO-8859-1 не прочитан: %+v, %v", got, err)
	}
}
