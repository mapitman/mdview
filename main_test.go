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

const testImageDataURI = "data:image/png;base64,aW1hZ2UgZGF0YQ=="

func newImageDir(t *testing.T, names ...string) string {
	t.Helper()
	baseDir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(baseDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("image data"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return baseDir
}

func render(t *testing.T, markdown string, baseDir string) string {
	t.Helper()
	var rendered bytes.Buffer
	if err := newMarkdown(baseDir).Convert([]byte(markdown), &rendered); err != nil {
		t.Fatal(err)
	}
	return rendered.String()
}

func TestMarkdownImagesAreEmbedded(t *testing.T) {
	baseDir := newImageDir(t, "images/diagram.png", "my pic.png", "a(b).png")

	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{"double-quoted title", `![x](images/diagram.png "Architecture diagram")`, `title="Architecture diagram"`},
		{"single-quoted title", `![x](images/diagram.png 'Architecture diagram')`, `title="Architecture diagram"`},
		{"parenthesized title", `![x](images/diagram.png (Architecture diagram))`, `title="Architecture diagram"`},
		{"without title", `![x](images/diagram.png)`, ``},
		{"angle-bracket destination with space", `![x](<my pic.png>)`, ``},
		{"percent-encoded space", `![x](my%20pic.png)`, ``},
		{"parentheses in name", `![x](<a(b).png>)`, ``},
		{"fragment", `![x](images/diagram.png#frag)`, ``},
		{"query", `![x](images/diagram.png?v=2)`, ``},
		{"nested brackets in alt", `![a [b] c](images/diagram.png)`, `alt="a [b] c"`},
		{"reference style", "![x][img]\n\n[img]: images/diagram.png", ``},
		{"html img", `<img src="images/diagram.png" alt="x">`, ``},
		{"html img single quotes", `<img src='images/diagram.png'>`, ``},
		{"html img unquoted", `<img src=images/diagram.png>`, ``},
		{"html img spaces around equals", `<img alt="x" src = "images/diagram.png">`, ``},
		{"html img block", "<p>\n<img\n  src=\"images/diagram.png\">\n</p>", ``},
		{"html img srcset", `<img src="images/diagram.png" srcset="images/diagram.png 2x">`, ``},
		{"html img > in attribute", `<img alt="a > b" src="images/diagram.png">`, ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render(t, tt.markdown, baseDir)
			if !strings.Contains(out, testImageDataURI) {
				t.Fatalf("image was not embedded: %s", out)
			}
			if tt.want != "" && !strings.Contains(out, tt.want) {
				t.Fatalf("want %s in: %s", tt.want, out)
			}
		})
	}
}

func TestSrcsetKeepsDescriptors(t *testing.T) {
	baseDir := newImageDir(t, "a.png", "b.png")
	out := render(t, `<img srcset="a.png 1x, b.png 2x">`, baseDir)
	if !strings.Contains(out, testImageDataURI+" 1x, "+testImageDataURI+" 2x") {
		t.Fatalf("srcset not rewritten as expected: %s", out)
	}
}

func TestImagesInCodeAreNotEmbedded(t *testing.T) {
	baseDir := newImageDir(t, "diagram.png")
	markdown := "`![x](diagram.png)`\n\n```\n![x](diagram.png)\n<img src=\"diagram.png\">\n```\n"
	if out := render(t, markdown, baseDir); strings.Contains(out, "base64") {
		t.Fatalf("image in code was embedded: %s", out)
	}
}

func TestNonEmbeddableImagesAreLeftAlone(t *testing.T) {
	baseDir := newImageDir(t, "notes.txt")
	for _, markdown := range []string{
		`![x](https://example.com/a.png)`,
		`![x](//example.com/a.png)`,
		`![x](file:///etc/hosts)`,
		`![x](ftp://example.com/a.png)`,
		`![x](blob:abc)`,
		`![x](missing.png)`,
		`![x](notes.txt)`,
		`![x](#frag)`,
	} {
		if out := render(t, markdown, baseDir); strings.Contains(out, "base64") {
			t.Errorf("%s was embedded: %s", markdown, out)
		}
	}
}

func TestGetMimeType(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n")
	tests := []struct {
		path string
		data []byte
		want string
	}{
		{"a.PNG", nil, "image/png"},
		{"a.avif", nil, "image/avif"},
		{"a.tif", nil, "image/tiff"},
		{"a.tiff", nil, "image/tiff"},
		{"a.svg", nil, "image/svg+xml"},
		{"noext", png, "image/png"},
		{"a.dat", png, "image/png"},
		{"a.txt", []byte("hello"), ""},
	}
	for _, tt := range tests {
		if got := getMimeType(tt.path, tt.data); got != tt.want {
			t.Errorf("getMimeType(%q) = %q, want %q", tt.path, got, tt.want)
		}
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
