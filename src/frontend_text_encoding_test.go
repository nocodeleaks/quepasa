package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUITextSourcesDoNotContainKnownMojibake(t *testing.T) {
	t.Parallel()

	roots := []string{
		filepath.Join("frontend", "src"),
		"views",
	}

	extensions := map[string]bool{
		".vue":  true,
		".ts":   true,
		".html": true,
		".tmpl": true,
		".css":  true,
	}

	mojibakeMarkers := []string{
		"ÃƒÂ¡",
		"ÃƒÂ¢",
		"ÃƒÂ£",
		"ÃƒÂ§",
		"ÃƒÂ©",
		"ÃƒÂª",
		"ÃƒÂ­",
		"ÃƒÂ³",
		"ÃƒÂ´",
		"ÃƒÂµ",
		"ÃƒÂº",
		"Ãƒâ€°",
		"Ãƒâ€¡",
		"ÃƒÆ’",
		"Ã‚ ",
		"Ã¢â‚¬",
		"Ã¢â‚¬â„¢",
		"Ã¢â‚¬Å“",
		"Ã¢â‚¬\u009d",
		"Ã¢â‚¬â€",
		"Ã¢â‚¬â€œ",
		"Ã¢â‚¬Â¢",
		"clÃ¡ssica",
		"sessÃ£o",
		"mÃºltiplas",
		"conexÃµes",
		"integraÃ§Ãµes",
		"FormulÃ¡rio",
		"Â©",
	}

	var findings []string
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if !extensions[filepath.Ext(path)] {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			text := string(content)
			for _, marker := range mojibakeMarkers {
				if strings.Contains(text, marker) {
					findings = append(findings, fmt.Sprintf("%s contains mojibake marker %q", path, marker))
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}

	if len(findings) > 0 {
		t.Fatalf("UI source files contain mojibake:\n%s", strings.Join(findings, "\n"))
	}
}
