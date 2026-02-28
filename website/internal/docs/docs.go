package docs

import (
	"embed"
	"errors"
	"path"
	"strings"
)

//go:embed html/*
var fs embed.FS

// LoadHTML returns an embedded docs snippet by slug (without ".html" suffix).
func LoadHTML(slug string) (string, error) {
	cleanSlug := strings.TrimSpace(slug)
	if cleanSlug == "" {
		return "", errors.New("docs slug is empty")
	}
	if strings.Contains(cleanSlug, "/") || strings.Contains(cleanSlug, "..") {
		return "", errors.New("invalid docs slug")
	}

	filePath := path.Join("html", cleanSlug+".html")
	content, err := fs.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
