package maven

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// property хранит произвольный именованный элемент секции properties.
type property struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// pom читает только параметры модели, без зависимостей плагинов и профилей.
type pom struct {
	XMLName xml.Name `xml:"project"`
	Coordinates
	Parent       *Coordinates `xml:"parent"`
	Properties   []property
	Dependencies []Dependency `xml:"dependencies>dependency"`
	Managed      []Dependency `xml:"dependencyManagement>dependencies>dependency"`
}

// propertiesXML позволяет читать произвольные имена свойств через encoding/xml.
type propertiesXML struct {
	Values []property `xml:",any"`
}

// decodePOM проверяет XML и извлекает только непосредственные поля project.
func decodePOM(data []byte) (pom, error) {
	var raw struct {
		XMLName xml.Name `xml:"project"`
		Coordinates
		Parent       *Coordinates  `xml:"parent"`
		Properties   propertiesXML `xml:"properties"`
		Dependencies []Dependency  `xml:"dependencies>dependency"`
		Managed      []Dependency  `xml:"dependencyManagement>dependencies>dependency"`
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charsetReader
	if err := decoder.Decode(&raw); err != nil {
		return pom{}, fmt.Errorf("некорректный XML POM: %w", err)
	}
	if err := checkXMLTail(decoder); err != nil {
		return pom{}, err
	}
	return pom{raw.XMLName, raw.Coordinates, raw.Parent, raw.Properties.Values,
		raw.Dependencies, raw.Managed}, nil
}

// checkXMLTail запрещает второй корневой элемент и текст после project.
func checkXMLTail(decoder *xml.Decoder) error {
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("некорректный конец XML POM: %w", err)
		}
		switch value := token.(type) {
		case xml.Comment, xml.ProcInst:
			continue
		case xml.CharData:
			if strings.TrimSpace(string(value)) == "" {
				continue
			}
		}
		return fmt.Errorf("лишние данные после корневого элемента POM")
	}
}

// charsetReader поддерживает ISO-8859-1, встречающийся в родительских POM Maven Central.
// Внешние библиотеки перекодировки не используются.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "iso-8859-1", "latin1":
		data, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}
		var text strings.Builder
		for _, value := range data {
			text.WriteRune(rune(value))
		}
		return strings.NewReader(text.String()), nil
	default:
		return nil, fmt.Errorf("неподдерживаемая кодировка POM %q", charset)
	}
}
