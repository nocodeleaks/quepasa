package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {
	// Read the qp_defaults.go file
	defaultsFile := "src/models/qp_defaults.go"
	content, err := os.ReadFile(defaultsFile)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", defaultsFile, err)
		os.Exit(1)
	}

	// Extract version using regex
	versionRegex := regexp.MustCompile(`const QpVersion = "([^"]+)"`)
	matches := versionRegex.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		fmt.Println("Could not find QpVersion constant")
		os.Exit(1)
	}

	version := matches[1]
	fmt.Printf("Found version: %s\n", version)

	// Read README.md
	readmeFile := "README.md"
	readmeContent, err := os.ReadFile(readmeFile)
	if err != nil {
		fmt.Printf("Error reading README.md: %v\n", err)
		os.Exit(1)
	}

	// Replace version placeholder in README
	// Look for a pattern like: <!-- VERSION: x.x.xxxx.xxxx -->
	versionPlaceholder := regexp.MustCompile(`<!-- VERSION: [^>]+ -->`)
	newVersionComment := fmt.Sprintf("<!-- VERSION: %s -->", version)

	readmeString := string(readmeContent)
	if versionPlaceholder.MatchString(readmeString) {
		readmeString = versionPlaceholder.ReplaceAllString(readmeString, newVersionComment)
		fmt.Println("Updated existing version comment")
	} else {
		// If no placeholder exists, add one at the top
		readmeString = newVersionComment + "\n" + readmeString
		fmt.Println("Added version comment at the top")
	}

	// Look for version display patterns and update them
	// Pattern 1: **Current Version:** `x.x.xxxx.xxxx`
	versionDisplayRegex := regexp.MustCompile(`\*\*Current Version:\*\* ` + "`" + `[\d.]+\.[\d.]+\.[\d.]+\.[\d.]+` + "`")
	if versionDisplayRegex.MatchString(readmeString) {
		readmeString = versionDisplayRegex.ReplaceAllString(readmeString, fmt.Sprintf("**Current Version:** `%s`", version))
		fmt.Println("Updated version display")
	} else {
		// Add version info after the title if not found
		titleRegex := regexp.MustCompile(`(# QuePasa\n\n> A micro web-application to make web-based WhatsApp bots easy to write\.)`)
		if titleRegex.MatchString(readmeString) {
			readmeString = titleRegex.ReplaceAllString(readmeString, fmt.Sprintf("$1\n\n**Current Version:** `%s`", version))
			fmt.Println("Added version info after title")
		}
	}

	// Write back to README.md
	err = os.WriteFile(readmeFile, []byte(readmeString), 0644)
	if err != nil {
		fmt.Printf("Error writing README.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully updated README.md with version %s\n", version)
}
