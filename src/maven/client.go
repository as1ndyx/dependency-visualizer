package maven

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// RequestTimeout ограничивает ожидание HTTP-ответа одного POM.
	RequestTimeout = 15 * time.Second
	// MaxPOMBytes ограничивает размер XML, загружаемого в память.
	MaxPOMBytes = 2 * 1024 * 1024
	// MaxModels ограничивает число родительских POM и BOM за один анализ.
	MaxModels = 64
)

// Client загружает POM через HTTP. Для тестов можно передать собственный HTTP-клиент.
type Client struct {
	http *http.Client
}

// NewClient создаёт клиент с тайм-аутом, если httpClient не передан.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: RequestTimeout}
	}
	return &Client{http: httpClient}
}

// POMURL строит адрес по схеме Maven: group/path/artifact/version/artifact-version.pom.
func POMURL(repository string, coordinates Coordinates) (string, error) {
	if _, err := ParseCoordinates(coordinates.String()); err != nil {
		return "", err
	}
	if strings.HasSuffix(coordinates.Version, "-SNAPSHOT") {
		return "", fmt.Errorf("SNAPSHOT-версии с maven-metadata.xml пока не поддерживаются")
	}
	u, err := url.Parse(repository)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("repository: ожидается абсолютный URL HTTP(S)")
	}
	if u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", fmt.Errorf("repository: URL содержит неподдерживаемые компоненты")
	}
	path := strings.ReplaceAll(coordinates.GroupID, ".", "/") + "/" + coordinates.ArtifactID
	path += "/" + coordinates.Version + "/" + coordinates.ArtifactID + "-" + coordinates.Version + ".pom"
	return strings.TrimRight(u.String(), "/") + "/" + path, nil
}

// fetch загружает ограниченный по размеру POM и проверяет статус HTTP.
func (c *Client) fetch(ctx context.Context, address string) (pom, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return pom{}, err
	}
	request.Header.Set("Accept", "application/xml, text/xml")
	response, err := c.http.Do(request)
	if err != nil {
		return pom{}, fmt.Errorf("не удалось получить POM %s: %w", address, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return pom{}, fmt.Errorf("POM %s: HTTP %d", address, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, MaxPOMBytes+1))
	if err != nil {
		return pom{}, fmt.Errorf("ошибка чтения POM %s: %w", address, err)
	}
	if len(data) > MaxPOMBytes {
		return pom{}, fmt.Errorf("POM %s превышает допустимый размер %d байт", address, MaxPOMBytes)
	}
	model, err := decodePOM(data)
	if err != nil {
		return pom{}, fmt.Errorf("POM %s: %w", address, err)
	}
	return model, nil
}
