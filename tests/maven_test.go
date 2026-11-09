package tests

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"dependencyvisualizer/src/maven"
)

func TestMavenPOMURL(t *testing.T) {
	coordinates, err := maven.ParseCoordinates("org.example:lib:1.2+build")
	if err != nil {
		t.Fatal(err)
	}
	for _, base := range []string{"https://example.org/maven2", "https://example.org/maven2/"} {
		got, err := maven.POMURL(base, coordinates)
		want := "https://example.org/maven2/org/example/lib/1.2+build/lib-1.2+build.pom"
		if err != nil || got != want {
			t.Fatalf("получили %q, ошибка %v", got, err)
		}
	}
}

func TestMavenDirectOnly(t *testing.T) {
	body := `<dependencies>` + dependencyXML("runtime", "2", "") +
		dependencyXML("tests", "3", `<scope>test</scope><optional>true</optional>`) + `</dependencies>`
	body += `<dependencyManagement><dependencies>` + dependencyXML("managed-only", "4", "") +
		`</dependencies></dependencyManagement><build><plugins><plugin><dependencies>` +
		dependencyXML("plugin-only", "5", "") + `</dependencies></plugin></plugins></build>`
	body += `<profiles><profile><dependencies>` + dependencyXML("profile-only", "6", "") +
		`</dependencies></profile></profiles>`
	server, requests := mavenServer(t, map[string]string{mavenPath("app"): pomXML(body)})
	got, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
	if err != nil {
		t.Fatal(err)
	}
	want := []maven.Dependency{
		{Coordinates: maven.Coordinates{GroupID: "org.example", ArtifactID: "runtime", Version: "2"},
			Type: "jar", Scope: "compile", Optional: "false"},
		{Coordinates: maven.Coordinates{GroupID: "org.example", ArtifactID: "tests", Version: "3"},
			Type: "jar", Scope: "test", Optional: "true"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %+v, ожидали %+v", got, want)
	}
	if counts := requests(); len(counts) != 1 || counts[mavenPath("app")] != 1 {
		t.Fatalf("лишние запросы, включая транзитивные: %v", counts)
	}
}

func TestMavenParentPropertiesAndBOM(t *testing.T) {
	parent := `<properties><lib.version>2</lib.version></properties><dependencies>` +
		dependencyXML("inherited", "${lib.version}", "") + `</dependencies>`
	root := parentXML("parent") + `<properties><lib.version>3</lib.version>` +
		`<alias>${lib.version}</alias></properties><dependencyManagement><dependencies>` +
		dependencyXML("bom", "1", `<type>pom</type><scope>import</scope>`) +
		dependencyXML("override", "${alias}", "") + `</dependencies></dependencyManagement><dependencies>` +
		dependencyXML("managed", "", "") + dependencyXML("override", "", "") +
		dependencyXML("self-version", "${project.version}", "") + `</dependencies>`
	bom := `<properties><managed.version>4</managed.version></properties><dependencyManagement><dependencies>` +
		dependencyXML("managed", "${managed.version}", "") + dependencyXML("override", "9", "") +
		`</dependencies></dependencyManagement>`
	server, requests := mavenServer(t, map[string]string{
		mavenPath("app"): pomXML(root), mavenPath("parent"): pomXML(parent), mavenPath("bom"): pomXML(bom),
	})
	got, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
	if err != nil {
		t.Fatal(err)
	}
	var versions []string
	for _, dependency := range got {
		versions = append(versions, dependency.Coordinates.String())
	}
	want := []string{"org.example:managed:4", "org.example:override:3", "org.example:self-version:1",
		"org.example:inherited:3"}
	if !reflect.DeepEqual(versions, want) {
		t.Fatalf("получили %v, ожидали %v", versions, want)
	}
	for path, count := range requests() {
		if count != 1 {
			t.Errorf("POM %s загружен %d раз", path, count)
		}
	}
}

func TestMavenEmptyAndOutput(t *testing.T) {
	server, _ := mavenServer(t, map[string]string{mavenPath("app"): pomXML("")})
	got, err := maven.NewClient(nil).DirectDependencies(context.Background(), server.URL, "org.example:app:1")
	if err != nil || len(got) != 0 {
		t.Fatalf("ожидался пустой список: %+v, %v", got, err)
	}
	var out strings.Builder
	if err := maven.PrintDependencies(&out, "org.example:app:1", got); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "нет прямых зависимостей") {
		t.Fatal(out.String())
	}
}

func TestMavenOutputMetadata(t *testing.T) {
	dependency := maven.Dependency{Coordinates: maven.Coordinates{GroupID: "org.example", ArtifactID: "lib", Version: "1"},
		Scope: "provided", Type: "test-jar", Classifier: "tests", Optional: "true"}
	var out strings.Builder
	if err := maven.PrintDependencies(&out, "org.example:app:1", []maven.Dependency{dependency}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"org.example:lib:1", "scope=provided", "type=test-jar",
		"classifier=tests", "optional=true"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("нет %q в %q", want, out.String())
		}
	}
}
