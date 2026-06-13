package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	rootDir := "."
	
	// 1. Rename files first
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip git, docker, node_modules etc.
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == ".gemini") {
			return filepath.SkipDir
		}
		
		if !d.IsDir() {
			newName := path
			newName = strings.Replace(newName, "seller", "seller", -1)
			newName = strings.Replace(newName, "note", "note", -1)
			newName = strings.Replace(newName, "seller_dashboard", "seller_dashboard", -1)
			
			if newName != path {
				fmt.Printf("Renaming %s to %s\n", path, newName)
				os.Rename(path, newName)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking directories for rename:", err)
	}

	// 2. Replace content
	extensions := map[string]bool{
		".go": true, ".sql": true, ".html": true, ".js": true, ".css": true, ".md": true,
	}

	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == ".gemini") {
			return filepath.SkipDir
		}
		
		if !d.IsDir() {
			ext := filepath.Ext(path)
			if extensions[ext] {
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				text := string(content)
				
				// Perform replacements
				// Order matters! Plurals and specific cases first.
				
				// specific variables/paths
				text = strings.Replace(text, "seller_dashboard", "seller_dashboard", -1)
				
				// SELLER -> SELLER
				text = strings.Replace(text, "SELLER", "SELLER", -1)
				text = strings.Replace(text, "Seller", "Seller", -1)
				text = strings.Replace(text, "seller", "seller", -1)
				
				// NOTES -> NOTES
				text = strings.Replace(text, "NOTES", "NOTES", -1)
				text = strings.Replace(text, "Notes", "Notes", -1)
				text = strings.Replace(text, "notes", "notes", -1)
				
				// NOTE -> NOTE
				text = strings.Replace(text, "NOTE", "NOTE", -1)
				text = strings.Replace(text, "Note", "Note", -1)
				text = strings.Replace(text, "note", "note", -1)
				
				
				if string(content) != text {
					fmt.Printf("Updating content in %s\n", path)
					os.WriteFile(path, []byte(text), 0644)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking directories for content replacement:", err)
	}
}
