package main

import (
	"fmt"
	"os"

	"github.com/mabrarov/go-equal-readers/cmp"
)

const bufSize = 4096

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: fcmp <file1> <file2>")
		os.Exit(1)
	}
	name1 := os.Args[1]
	name2 := os.Args[2]
	f1, err := os.Open(name1)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", name1, err)
		os.Exit(2)
	}
	defer func() { _ = f1.Close() }()
	f2, err := os.Open(name2)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", name2, err)
		os.Exit(2)
	}
	defer func() { _ = f2.Close() }()
	buf1 := make([]byte, bufSize)
	buf2 := make([]byte, bufSize)
	eq, err := cmp.EqualReaders(buf1, buf2, 2, f1, f2)
	if err != nil {
		fmt.Printf("Failed to compare files: %v\n", err)
		os.Exit(2)
	}
	if eq {
		fmt.Printf("File %s is equal to file %s\n", name1, name2)
	} else {
		fmt.Printf("File %s is NOT equal to file %s\n", name1, name2)
	}
}
