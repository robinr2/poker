package main

import (
	"fmt"
	"os"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Printf("Working directory: %s\n", wd)
	
	paths := []string{
		"web/static/index.html",
	}
	
	for _, p := range paths {
		if stat, err := os.Stat(p); err != nil {
			fmt.Printf("NOT FOUND: %s - %v\n", p, err)
		} else {
			fmt.Printf("FOUND: %s (size: %d)\n", p, stat.Size())
		}
	}
}
