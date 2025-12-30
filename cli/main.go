package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: go run cli/main.go wiki <title>")
		os.Exit(1)
	}

	contentType := os.Args[1]
	title := os.Args[2]

	if contentType != "wiki" {
		fmt.Fprintln(os.Stderr, "Invalid type. Must be 'wiki'.")
		os.Exit(1)
	}

	slug := createSlug(title)
	fileName := slug + ".md"

	var dir string
	if contentType == "blog" {
		dir = "blogs"
	} else {
		dir = "wikis"
	}

	filePath := filepath.Join(dir, fileName)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "File already exists: %s\n", filePath)
		os.Exit(1)
	}

	currentTime := time.Now().UTC().Format(time.RFC3339)

	content := fmt.Sprintf(`---
title: "%s"
desc: ""
createdAt: "%s"
---
`, title, currentTime)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created new %s: %s\n", contentType, filePath)
}

func createSlug(title string) string {
	slug := strings.ToLower(title)
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
