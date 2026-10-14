package main

import (
	"bytes"
	"flag"
	"fmt"
	stdhtml "html"
	"os"
	"path/filepath"
	"reflect"
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
	tests := map[string]string{
		"a.PNG":  "image/png",
		"a.avif": "image/avif",
		"a.tif":  "image/tiff",
		"a.tiff": "image/tiff",
		"a.svg":  "image/svg+xml",
		"noext":  "",
		"a.dat":  "",
		"a.txt":  "",
		"id_rsa": "",
	}
	for path, want := range tests {
		if got := getMimeType(path); got != want {
			t.Errorf("getMimeType(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestImageToDataURIPathRules(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "a", "b", "c", "d", "docs")
	write := func(path string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("image data"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(base, "in.png"))
	write(filepath.Join(base, "sub", "in.png"))
	write(filepath.Join(base, "..", "up1.png"))
	write(filepath.Join(base, "..", "..", "..", "up3.png"))
	write(filepath.Join(base, "..", "..", "..", "..", "up4.png"))
	write(filepath.Join(base+"-secret", "x.png")) // sibling that shares the name prefix
	write(filepath.Join(base, "key.pem"))
	write(filepath.Join(root, "outside", "target.png"))

	if err := os.Symlink(filepath.Join(root, "outside", "target.png"), filepath.Join(base, "link.png")); err != nil {
		t.Skip("symlinks not supported:", err)
	}
	if err := os.Symlink(filepath.Join(base, "in.png"), filepath.Join(base, "inlink.png")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(base, "key.pem"), filepath.Join(base, "disguised.png")); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path  string
		embed bool
	}{
		{"in.png", true},
		{"sub/in.png", true},
		{"../up1.png", true},
		{"../../../up3.png", true},
		{"../../../../up4.png", false},
		{"../docs-secret/x.png", true}, // one level up is allowed
		{"key.pem", false},
		{"missing.png", false},
		{"inlink.png", true},
		{"link.png", false}, // symlink to a file outside the allowed area
		{"sub", false},      // directory
	}
	for _, tt := range tests {
		got := imageToDataURI(tt.path, base) != ""
		if got != tt.embed {
			t.Errorf("imageToDataURI(%q) embedded = %v, want %v", tt.path, got, tt.embed)
		}
	}
}

func TestImageToDataURISizeLimit(t *testing.T) {
	base := t.TempDir()
	atLimit := filepath.Join(base, "limit.png")
	tooBig := filepath.Join(base, "big.png")
	if err := os.WriteFile(atLimit, make([]byte, maxImageSize), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tooBig, make([]byte, maxImageSize+1), 0644); err != nil {
		t.Fatal(err)
	}
	if imageToDataURI("limit.png", base) == "" {
		t.Error("file at the size limit was not embedded")
	}
	if imageToDataURI("big.png", base) != "" {
		t.Error("file over the size limit was embedded")
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

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantPos  []string
		wantOut  string
		wantBare bool
	}{
		{"no arguments", nil, nil, "", false},
		{"flags first", []string{"-o", "x.html", "a.md"}, []string{"a.md"}, "x.html", false},
		{"flags after filename", []string{"a.md", "-o", "x.html", "-b"}, []string{"a.md"}, "x.html", true},
		{"flags on both sides", []string{"-b", "a.md", "-o", "x.html"}, []string{"a.md"}, "x.html", true},
		{"stdin marker", []string{"-", "-o", "x.html"}, []string{"-"}, "x.html", false},
		{"double dash", []string{"-b", "--", "-weird.md"}, []string{"-weird.md"}, "", true},
		{"flags after double dash are positional", []string{"--", "a.md", "-b"}, []string{"a.md", "-b"}, "", false},
		{"several files", []string{"a.md", "b.md"}, []string{"a.md", "b.md"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("mdview", flag.ContinueOnError)
			out := fs.String("o", "", "")
			bare := fs.Bool("b", false, "")

			got, err := parseFlags(fs, tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.wantPos) {
				t.Errorf("positional = %q, want %q", got, tt.wantPos)
			}
			if *out != tt.wantOut || *bare != tt.wantBare {
				t.Errorf("o = %q, b = %v; want o = %q, b = %v", *out, *bare, tt.wantOut, tt.wantBare)
			}
		})
	}
}
