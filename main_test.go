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
	"time"

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

func TestHasMermaidBlock(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     bool
	}{
		{"mermaid fence", "```mermaid\ngraph TD; A-->B\n```\n", true},
		{"other fence", "```go\nfmt.Println()\n```\n", false},
		{"no code", "# Title\n\ntext\n", false},
		{"raw html mentioning mermaid", "<pre class=\"mermaid\">graph TD</pre>\n", false},
		{"code span mentioning mermaid", "Use `class=\"mermaid\"` in HTML.\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.markdown)
			doc := newMarkdown(".").Parser().Parse(text.NewReader(source))
			if got := hasMermaidBlock(doc); got != tt.want {
				t.Fatalf("hasMermaidBlock = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmbedMermaidScript(t *testing.T) {
	out := embedMermaidScript("<p>x</p>")
	if !strings.HasPrefix(out, "<p>x</p><script>") || !strings.Contains(out, "mermaid.initialize") {
		t.Fatalf("mermaid script was not appended: %.80s", out)
	}
	if strings.Count(out, "</script>") != 2 {
		t.Fatalf("expected exactly two closing script tags, got %d", strings.Count(out, "</script>"))
	}
}

func TestGetTempDir(t *testing.T) {
	t.Run("MDVIEW_DIR is created with parents", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "a", "b", "c")
		t.Setenv("MDVIEW_DIR", dir)
		if got := getTempDir(); got != dir {
			t.Fatalf("getTempDir = %q, want %q", got, dir)
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Fatalf("directory was not created: %v", err)
		}
	})

	t.Run("snap uses mdview-temp in HOME", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("MDVIEW_DIR", "")
		t.Setenv("SNAP_USER_COMMON", "/snap/common")
		t.Setenv("HOME", home)
		want := filepath.Join(home, "mdview-temp")
		if got := getTempDir(); got != want {
			t.Fatalf("getTempDir = %q, want %q", got, want)
		}
	})

	t.Run("default is the OS temp dir", func(t *testing.T) {
		t.Setenv("MDVIEW_DIR", "")
		t.Setenv("SNAP_USER_COMMON", "")
		if got := getTempDir(); got != os.TempDir() {
			t.Fatalf("getTempDir = %q, want %q", got, os.TempDir())
		}
	})
}

func TestWriteOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")
	if err := writeOutput(path, "body{}", "T", "<p>hi</p>"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(got), "<title>T</title>") || !strings.Contains(string(got), "<p>hi</p>") {
		t.Fatalf("unexpected output (err %v): %s", err, got)
	}
	if err := writeOutput(filepath.Join(dir, "missing", "out.html"), "", "T", ""); err == nil {
		t.Fatal("expected an error for a missing directory")
	}
}

func TestRemoveStaleTempFiles(t *testing.T) {
	dir := t.TempDir()
	name := func(c string) string { return "mdview" + strings.Repeat(c, 32) + ".html" }
	old := time.Now().Add(-48 * time.Hour)

	files := map[string]time.Time{
		name("a"):         old,        // stale generated file: removed
		name("b"):         time.Now(), // fresh generated file: kept
		"notes.html":      old,        // other file: kept
		"mdview-old.html": old,        // wrong name pattern: kept
		name("c") + "x":   old,        // wrong suffix: kept
	}
	for file, mtime := range files {
		path := filepath.Join(dir, file)
		if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}

	removeStaleTempFiles(dir, staleTempFileAge)

	for file := range files {
		_, err := os.Stat(filepath.Join(dir, file))
		wantRemoved := file == name("a")
		if wantRemoved != os.IsNotExist(err) {
			t.Errorf("%s: removed = %v, want %v", file, os.IsNotExist(err), wantRemoved)
		}
	}
}

func TestGetTitleFromAST(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{"simple", "# Hello\n", "Hello"},
		{"first H1 wins", "# First\n\n# Second\n", "First"},
		{"skips other levels", "## Sub\n\n# Main\n", "Main"},
		{"no H1", "## Sub\n\ntext\n", ""},
		{"empty document", "", ""},
		{"inline markup", "# Hello *big* **bold** `code` [link](http://x)\n", "Hello big bold code link"},
		{"setext heading", "Title\n=====\n", "Title"},
		{"surrounding space", "#   Spaced   \n", "Spaced"},
		{"H1 below a paragraph", "text\n\n# Later\n", "Later"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.markdown)
			doc := goldmark.New().Parser().Parse(text.NewReader(source))
			if got := getTitleFromAST(doc, source); got != tt.want {
				t.Fatalf("getTitleFromAST = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsRelativePath(t *testing.T) {
	tests := map[string]bool{
		"image.png":                true,
		"./image.png":              true,
		"../image.png":             true,
		"dir/image.png":            true,
		"my image.png":             true,
		"":                         false,
		"/etc/image.png":           false,
		"//cdn.example.com/a.png":  false,
		"http://example.com/a.png": false,
		"https://example.com/a":    false,
		"data:image/png;base64,AA": false,
		"file:///tmp/a.png":        false,
		"ftp://example.com/a.png":  false,
		"blob:abc":                 false,
		"mailto:a@b.c":             false,
	}
	for path, want := range tests {
		if got := isRelativePath(path); got != want {
			t.Errorf("isRelativePath(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestTempFileName(t *testing.T) {
	dir := t.TempDir()
	first := tempFileName(dir, "mdview", ".html")
	second := tempFileName(dir, "mdview", ".html")

	if first == second {
		t.Fatal("two temp file names were identical")
	}
	if filepath.Dir(first) != dir {
		t.Fatalf("temp file %q is not in %q", first, dir)
	}
	if !tempFileRegex.MatchString(filepath.Base(first)) {
		t.Fatalf("temp file name %q does not match the cleanup pattern", filepath.Base(first))
	}
}
