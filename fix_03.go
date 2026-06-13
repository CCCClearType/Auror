package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	path := "db/03_init_super_data.sql"
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}
	text := string(content)

	// Refactor like before
	text = strings.Replace(text, "DEVELOPER", "SELLER", -1)
	text = strings.Replace(text, "Developer", "Seller", -1)
	text = strings.Replace(text, "developer", "seller", -1)
	text = strings.Replace(text, "GAMES", "NOTES", -1)
	text = strings.Replace(text, "Games", "Notes", -1)
	text = strings.Replace(text, "games", "notes", -1)
	text = strings.Replace(text, "GAME", "NOTE", -1)
	text = strings.Replace(text, "Game", "Note", -1)
	text = strings.Replace(text, "game", "note", -1)

	// Fix duplicate tags
	text = strings.Replace(text, "(100, '資料結構')", "(100, '資料結構(高階)')", -1)
	text = strings.Replace(text, "(101, '演算法')", "(101, '演算法(高階)')", -1)
	text = strings.Replace(text, "(102, '材料與生活')", "(102, '材料與生活(進階)')", -1)
	text = strings.Replace(text, "(103, '微積分')", "(103, '微積分(高階)')", -1)

	err = os.WriteFile(path, []byte(text), 0644)
	if err != nil {
		fmt.Println("Error writing:", err)
	} else {
		fmt.Println("Fixed 03_init_super_data.sql successfully!")
	}
}
