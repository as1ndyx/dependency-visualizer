package tests

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"dependencyvisualizer/src/testrepo"
)

func TestTestRepositoryParse(t *testing.T) {
	text := "\uFEFF# комментарий\r\n; комментарий\r\n\r\nA: B C B\r\nB: C\r\nC:\r\n"
	got, err := testrepo.Parse(strings.NewReader(text))
	want := testrepo.Repository{"A": {"B", "C"}, "B": {"C"}, "C": nil}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %+v, %v", got, err)
	}
	dependencies, err := got.Dependencies(context.Background(), "A")
	if err != nil {
		t.Fatal(err)
	}
	dependencies[0] = "CHANGED"
	if got["A"][0] != "B" {
		t.Fatal("источник позволяет изменить данные через возвращённый срез")
	}
}

func TestTestRepositoryErrors(t *testing.T) {
	cases := []struct{ text, want string }{
		{"", "не содержит"}, {"# comment\n", "не содержит"}, {"A B", "строка 1"},
		{"a:", "латинскими"}, {"А:", "латинскими"}, {"A1:", "латинскими"},
		{"A: b", "имя зависимости"}, {"A: B,C", "имя зависимости"},
		{"A: B: C", "имя зависимости"}, {"A:\nA:", "повторно"},
		{"A: B", "необъявленный пакет B"}, {"A:\nX: Y", "необъявленный пакет Y"},
		{strings.Repeat("x", testrepo.MaxFileBytes+1), "превышает"},
	}
	for _, tc := range cases {
		_, err := testrepo.Parse(strings.NewReader(tc.text))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("ожидали %q, получили %v", tc.want, err)
		}
	}
}

func TestTestRepositoryAllowsCycles(t *testing.T) {
	for _, text := range []string{"A: A", "A: B\nB: A"} {
		if _, err := testrepo.Parse(strings.NewReader(text)); err != nil {
			t.Fatal(err)
		}
	}
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, fmt.Errorf("ошибка чтения") }

func TestTestRepositoryIOErrors(t *testing.T) {
	if _, err := testrepo.Parse(brokenReader{}); err == nil {
		t.Fatal("потеряна ошибка чтения")
	}
	if _, err := testrepo.Load("missing-repository.txt"); err == nil {
		t.Fatal("принят отсутствующий файл")
	}
	repository := testrepo.Repository{"A": nil}
	if _, err := repository.Dependencies(context.Background(), "B"); err == nil {
		t.Fatal("принят неизвестный пакет")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repository.Dependencies(ctx, "A"); err != context.Canceled {
		t.Fatalf("отмена: %v", err)
	}
}
