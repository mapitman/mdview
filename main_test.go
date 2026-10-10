package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
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
