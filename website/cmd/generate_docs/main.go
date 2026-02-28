package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

const (
	SRC_DIR string = "docs"
	OUT_DIR string = "internal/docs/html"
)

func main() {
	// Recreate directory to get rid of old files
	_ = os.RemoveAll(OUT_DIR)
	if err := os.MkdirAll(OUT_DIR, os.ModePerm); err != nil {
		panic(err)
	}

	mdParser := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
				highlighting.WithGuessLanguage(true),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(false),
					chromahtml.WithLineNumbers(true),
					chromahtml.TabWidth(2),
				),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
	)

	files, err := os.ReadDir(SRC_DIR)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		mdData, err := os.ReadFile(filepath.Join(SRC_DIR, file.Name()))
		if err != nil {
			panic(err)
		}

		var bodyBuf bytes.Buffer
		if err := mdParser.Convert(mdData, &bodyBuf); err != nil {
			panic(err)
		}

		snippet := wrapSnippet(bodyBuf.String())

		htmlFilename := strings.TrimSuffix(file.Name(), ".md") + ".html"
		outPath := filepath.Join(OUT_DIR, htmlFilename)
		if err := os.WriteFile(outPath, []byte(snippet), 0644); err != nil {
			panic(err)
		}
		fmt.Printf("✅ Generated: %s\n", outPath)
	}
}

func wrapSnippet(body string) string {
	return `<article class="prose max-w-none docs-markdown">` + "\n" + body + "\n</article>\n"
}
