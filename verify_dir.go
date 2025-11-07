package main

import (
	"fmt"
	"os"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Printf("Working directory: %s\n", wd)
	
	if _, err := os.Stat("web/static/index.html"); err != nil {
		fmt.Printf("ERROR: web/static/index.html not found: %v\n", err)
	} else {
		fmt.Printf("OK: web/static/index.html found\n")
	}
}
