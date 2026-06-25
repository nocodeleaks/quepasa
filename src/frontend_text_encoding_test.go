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
		filepath.Join("apps", "vuejs", "client", "src"),
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
		"ÃƒÆ’Ã‚Â¡",
		"ÃƒÆ’Ã‚Â¢",
		"ÃƒÆ’Ã‚Â£",
		"ÃƒÆ’Ã‚Â§",
		"ÃƒÆ’Ã‚Â©",
		"ÃƒÆ’Ã‚Âª",
		"ÃƒÆ’Ã‚Â­",
		"ÃƒÆ’Ã‚Â³",
		"ÃƒÆ’Ã‚Â´",
		"ÃƒÆ’Ã‚Âµ",
		"ÃƒÆ’Ã‚Âº",
		"ÃƒÆ’Ã¢â‚¬Â°",
		"ÃƒÆ’Ã¢â‚¬Â¡",
		"ÃƒÆ’Ã†â€™",
		"Ãƒâ€š ",
		"ÃƒÂ¢Ã¢â€šÂ¬",
		"ÃƒÂ¢Ã¢â€šÂ¬Ã¢â€žÂ¢",
		"ÃƒÂ¢Ã¢â€šÂ¬Ã…â€œ",
		"ÃƒÂ¢Ã¢â€šÂ¬\u009d",
		"ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Â",
		"ÃƒÂ¢Ã¢â€šÂ¬Ã¢â‚¬Å“",
		"ÃƒÂ¢Ã¢â€šÂ¬Ã‚Â¢",
		"clÃƒÂ¡ssica",
		"sessÃƒÂ£o",
		"mÃƒÂºltiplas",
		"conexÃƒÂµes",
		"integraÃƒÂ§ÃƒÂµes",
		"FormulÃƒÂ¡rio",
		"Ã‚Â©",
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

// TestGoSourcesDoNotContainMojibake checks that Go source files do not contain
// single-level UTF-8-interpreted-as-Latin-1 mojibake sequences commonly introduced
// by editors that save files with the wrong encoding.
func TestGoSourcesDoNotContainMojibake(t *testing.T) {
	t.Parallel()

	// Single-level mojibake: a UTF-8 file was read as Latin-1 and re-saved as UTF-8.
	// Result: each accented char X (UTF-8 [0xC3, b]) becomes the 4-byte sequence
	// Ã (U+00C3 → [0xC3,0x83]) + <Latin-1 continuation char> ([0xC2, b]).
	// Markers are built from raw bytes so no literal mojibake sequence appears in
	// this source file (which would trigger a false self-match).
	simpleMojibakeMarkers := func() []string {
		// Second byte of the original UTF-8 two-byte sequence (0xC3, b).
		secondBytes := []byte{
			0xA7, // ç (U+00E7)
			0xA1, // á (U+00E1)
			0xA9, // é (U+00E9)
			0xAD, // í (U+00ED)
			0xB3, // ó (U+00F3)
			0xBA, // ú (U+00FA)
			0xA3, // ã (U+00E3)
			0xB5, // õ (U+00F5)
			0xA0, // à (U+00E0)
			0xA2, // â (U+00E2)
			0xAA, // ê (U+00EA)
			0xB4, // ô (U+00F4)
			0xB1, // ñ (U+00F1)
		}
		markers := make([]string, len(secondBytes))
		for i, b := range secondBytes {
			// Ã in UTF-8: [0xC3, 0x83]; Latin-1 continuation char: [0xC2, b]
			markers[i] = string([]byte{0xC3, 0x83, 0xC2, b})
		}
		return markers
	}()

	var findings []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			skip := d.Name() == "vendor" || d.Name() == "node_modules" || d.Name() == ".git"
			if skip {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		// Skip test files: marker literals inside test sources are not mojibake.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		text := string(content)
		for _, marker := range simpleMojibakeMarkers {
			if strings.Contains(text, marker) {
				findings = append(findings, fmt.Sprintf("%s contains mojibake marker %q", path, marker))
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	if len(findings) > 0 {
		t.Fatalf("Go source files contain mojibake:\n%s", strings.Join(findings, "\n"))
	}
}
