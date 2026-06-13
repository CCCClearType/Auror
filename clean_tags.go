package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	path := "db/03_init_super_data.sql"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	var newLines []string
	scanner := bufio.NewScanner(file)
	
	reTag := regexp.MustCompile(`INSERT INTO tags \(tag_id, tag_name\) VALUES \((\d+),`)
	reNoteTag := regexp.MustCompile(`INSERT INTO note_tags \(note_id, tag_id\) VALUES \(\d+,\s*(\d+)\);`)

	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.Contains(line, "INSERT INTO tags") {
			matches := reTag.FindStringSubmatch(line)
			if len(matches) > 1 {
				id, _ := strconv.Atoi(matches[1])
				if id >= 110 && id <= 179 {
					continue // skip this line
				}
			}
		} else if strings.Contains(line, "INSERT INTO note_tags") {
			matches := reNoteTag.FindStringSubmatch(line)
			if len(matches) > 1 {
				id, _ := strconv.Atoi(matches[1])
				if id >= 110 && id <= 179 {
					continue // skip this line
				}
			}
		}
		newLines = append(newLines, line)
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		fmt.Println("Scanner error:", err)
		return
	}

	err = os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
	if err != nil {
		fmt.Println("Error writing:", err)
	} else {
		fmt.Println("Successfully cleaned 03_init_super_data.sql")
	}
}
