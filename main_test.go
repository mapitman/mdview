package main

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func TestProcessMarkdownImages(t *testing.T) {
	baseDir := t.TempDir()
	imagePath := filepath.Join(baseDir, "images", "diagram.png")
	if err := os.MkdirAll(filepath.Dir(imagePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(imagePath, []byte("image data"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		markdown      string
		wantTitleHTML string
	}{
		{
			name:          "double-quoted title",
			markdown:      `![Architecture diagram](images/diagram.png "Architecture diagram")`,
			wantTitleHTML: `title="Architecture diagram"`,
		},
		{
			name:          "single-quoted title",
			markdown:      `![Architecture diagram](images/diagram.png 'Architecture diagram')`,
			wantTitleHTML: `title="Architecture diagram"`,
		},
		{
			name:          "parenthesized title",
			markdown:      `![Architecture diagram](images/diagram.png (Architecture diagram))`,
			wantTitleHTML: `title="Architecture diagram"`,
		},
		{
			name:     "without title",
			markdown: `![Architecture diagram](images/diagram.png)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processed := processMarkdownImages(tt.markdown, baseDir)
			if !strings.Contains(processed, `data:image/png;base64,aW1hZ2UgZGF0YQ==`) {
				t.Fatalf("image was not embedded: %s", processed)
			}

			var rendered bytes.Buffer
			if err := goldmark.New().Convert([]byte(processed), &rendered); err != nil {
				t.Fatal(err)
			}
			if tt.wantTitleHTML != "" && !strings.Contains(rendered.String(), tt.wantTitleHTML) {
				t.Fatalf("rendered image lost its title attribute: %s", rendered.String())
			}
		})
	}
}

func TestTitleIsEscaped(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"ampersand", "# Q&A\n", "<title>Q&amp;A</title>"},
		{"less-than", "# 1 < 2 > 0\n", "<title>1 &lt; 2 &gt; 0</title>"},
		{"entity", "# &lt;b&gt; &amp; c\n", "<title>&lt;b&gt; &amp; c</title>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.source)
			doc := goldmark.New().Parser().Parse(text.NewReader(source))
			title := getTitleFromAST(doc, source)

			var out bytes.Buffer
			fmt.Fprintf(&out, template, "", stdhtml.EscapeString(title), "")
			if !strings.Contains(out.String(), tt.want) {
				t.Fatalf("want %s in: %s", tt.want, out.String())
			}
		})
	}
}
